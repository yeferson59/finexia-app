// Package bold speaks to Bold, the Colombian payment gateway behind /apoyar:
// the payment button's integrity signature, the webhook's signature, and the
// payment-voucher query (developers.bold.co).
//
// It knows Bold's wire format and nothing about contributions: what a status
// means for an order, and what is stored, is the support module's business.
//
// Two keys, both from Bold's panel, each with a test and a production version:
//
//   - the identity key (API key) goes to the browser inside the checkout
//     options and authenticates the voucher query. Bold treats it as public.
//   - the secret key signs orders and webhooks. It never leaves the backend.
package bold

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	json "github.com/bytedance/sonic"
)

// Currency is the only one the contributions use.
const Currency = "COP"

// SignatureHeader carries the webhook's HMAC.
const SignatureHeader = "X-Bold-Signature"

const voucherURL = "https://payments.api.bold.co/v2/payment-voucher/"

// DefaultHTTPClient answers a person who is waiting on the page they came
// back to, so a slow Bold must give up long before they do.
var DefaultHTTPClient = new(http.Client{Timeout: 8 * time.Second})

// IntegritySignature is the hex SHA-256 of {orderID}{amount}{currency}{secret},
// in that order and with no separators, which is what the payment button checks
// before it lets anyone pay. It is what stops a browser from lowering the
// amount of an order it was handed.
func IntegritySignature(orderID string, amount int64, currency, secretKey string) string {
	sum := sha256.Sum256([]byte(orderID + strconv.FormatInt(amount, 10) + currency + secretKey))

	return hex.EncodeToString(sum[:])
}

// VerifyWebhook reports whether signature is Bold's HMAC of body: HMAC-SHA256,
// keyed with the secret key, over the body encoded in base64, as lowercase hex.
//
// body must be the raw bytes as received. Re-encoding a decoded payload would
// reorder or respace it and no signature would ever match.
//
// In Bold's test environment webhooks are signed with an empty key, so secret
// is "" there. That makes test-mode webhooks forgeable by anyone who knows the
// URL, which is why the composition root refuses test mode in production.
func VerifyWebhook(body []byte, signature, secret string) bool {
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(base64.StdEncoding.EncodeToString(body)))
	want := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(want), []byte(strings.ToLower(strings.TrimSpace(signature))))
}

// Event types of the webhook.
const (
	EventSaleApproved = "SALE_APPROVED"
	EventSaleRejected = "SALE_REJECTED"
	EventVoidApproved = "VOID_APPROVED"
	EventVoidRejected = "VOID_REJECTED"
)

// Event is the part of a webhook notification the app reads. The rest —
// payer's email, card, seller — is decoded by nobody.
type Event struct {
	// ID is the notification's own id; a retry of the same notification keeps it.
	ID   string
	Type string
	// PaymentID is Bold's id for the transaction (the notification's subject).
	PaymentID string
	// Reference is the order id the checkout was opened with, or "" for a
	// payment that did not come from one (a payment link, the dataphone).
	Reference     string
	Total         *int64
	PaymentMethod string
}

type eventPayload struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Subject string `json:"subject"`
	Data    struct {
		PaymentID string `json:"payment_id"`
		Amount    struct {
			Total *float64 `json:"total"`
		} `json:"amount"`
		PaymentMethod string `json:"payment_method"`
		Metadata      struct {
			Reference string `json:"reference"`
		} `json:"metadata"`
	} `json:"data"`
}

// ErrMalformedEvent is a body that is signed but not a notification this
// package can read.
var ErrMalformedEvent = errors.New("bold: malformed webhook event")

// ParseEvent decodes a webhook body. Call it only after VerifyWebhook.
func ParseEvent(body []byte) (Event, error) {
	var p eventPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return Event{}, fmt.Errorf("%w: %v", ErrMalformedEvent, err)
	}

	paymentID := p.Data.PaymentID
	if paymentID == "" {
		paymentID = p.Subject
	}

	if p.ID == "" || p.Type == "" || paymentID == "" {
		return Event{}, fmt.Errorf("%w: id, type and payment id are required", ErrMalformedEvent)
	}

	return Event{
		ID:            p.ID,
		Type:          p.Type,
		PaymentID:     paymentID,
		Reference:     p.Data.Metadata.Reference,
		Total:         pesos(p.Data.Amount.Total),
		PaymentMethod: p.Data.PaymentMethod,
	}, nil
}

// Statuses of the voucher query. PROCESSING and PENDING are not final; the
// last one means the order exists but nobody has paid it yet.
const (
	StatusApproved      = "APPROVED"
	StatusProcessing    = "PROCESSING"
	StatusPending       = "PENDING"
	StatusRejected      = "REJECTED"
	StatusFailed        = "FAILED"
	StatusVoided        = "VOIDED"
	StatusNoTransaction = "NO_TRANSACTION_FOUND"
)

// Payment is what the voucher query says about an order.
type Payment struct {
	Status        string
	Total         *int64
	TransactionID string
	PaymentMethod string
}

type voucherPayload struct {
	PaymentStatus string   `json:"payment_status"`
	Total         *float64 `json:"total"`
	TransactionID string   `json:"transaction_id"`
	PaymentMethod string   `json:"payment_method"`
}

// Client queries Bold with the identity key.
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

// New builds the client. A nil httpClient uses DefaultHTTPClient.
func New(httpClient *http.Client, apiKey string) *Client {
	if httpClient == nil {
		httpClient = DefaultHTTPClient
	}

	return new(Client{httpClient: httpClient, apiKey: apiKey, baseURL: voucherURL})
}

// Payment asks Bold for the state of an order,
// GET /v2/payment-voucher/{orderID}.
func (c *Client) Payment(ctx context.Context, orderID string) (Payment, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+url.PathEscape(orderID), nil)
	if err != nil {
		return Payment{}, err
	}

	req.Header.Set("Authorization", "x-api-key "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return Payment{}, fmt.Errorf("bold: payment voucher: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return Payment{}, fmt.Errorf("bold: payment voucher: status %d", res.StatusCode)
	}

	var p voucherPayload
	if err := json.ConfigDefault.NewDecoder(res.Body).Decode(&p); err != nil {
		return Payment{}, fmt.Errorf("bold: payment voucher: %w", err)
	}

	if p.PaymentStatus == "" {
		return Payment{}, errors.New("bold: payment voucher: no payment_status")
	}

	return Payment{
		Status:        p.PaymentStatus,
		Total:         pesos(p.Total),
		TransactionID: p.TransactionID,
		PaymentMethod: p.PaymentMethod,
	}, nil
}

// pesos turns Bold's JSON number into whole pesos. Bold sends COP without
// decimals, but as a JSON number that a strict integer decode would reject if
// it ever arrived as 10000.0.
func pesos(v *float64) *int64 {
	if v == nil || math.IsNaN(*v) || math.IsInf(*v, 0) {
		return nil
	}

	n := int64(math.Round(*v))

	return &n
}
