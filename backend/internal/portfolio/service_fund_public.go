package portfolio

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"uuid"
)

// The SFC's catalog of funds and the links to it; see fund_public.go.

// SearchPublicFunds finds funds of the SFC's catalog by name or code. A
// catalog nobody filled yet — a deployment the job has not run on — is filled
// on the spot, so the first search does not come back empty.
func (s *service) SearchPublicFunds(ctx context.Context, q string) ([]PublicFund, error) {
	words, err := publicFundSearchWords(q)
	if err != nil {
		return nil, err
	}

	n, err := s.repo.CountPublicFunds(ctx)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		if _, err := s.RefreshPublicFunds(ctx); err != nil {
			return nil, err
		}
	}

	since := cashRateDay(time.Now()).AddDate(0, 0, -publicFundStaleDays)

	return s.repo.SearchPublicFunds(ctx, words, since, publicFundSearchLimit)
}

// RefreshPublicFunds reads the latest day the SFC published into the catalog,
// and answers how many funds it wrote.
func (s *service) RefreshPublicFunds(ctx context.Context) (int, error) {
	if s.publicFunds == nil {
		return 0, ErrPublicFundsUnavailable
	}

	published, err := s.publicFunds.LatestFunds(ctx)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrPublicFundsUnavailable, err)
	}

	funds := make([]PublicFund, len(published))
	for i, f := range published {
		funds[i] = newPublicFund(f)
	}

	return s.repo.UpsertPublicFunds(ctx, funds)
}

// LinkFund links a fund followed by units to a fund of the SFC's catalog, and
// imports what it published since the fund's first purchase. The values are
// read before the write, so an SFC that does not answer leaves the fund as it
// was.
func (s *service) LinkFund(ctx context.Context, userID, assetID uuid.UUID, publicID string) (Fund, error) {
	fund, err := s.repo.GetFund(ctx, userID, assetID)
	if err != nil {
		return Fund{}, err
	}

	if err := fund.linkable(); err != nil {
		return Fund{}, err
	}

	public, err := s.repo.GetPublicFund(ctx, publicID)
	if err != nil {
		return Fund{}, err
	}

	movements, err := s.repo.GetFundMovements(ctx, userID, assetID)
	if err != nil {
		return Fund{}, err
	}

	from := cashRateDay(time.Now()).AddDate(0, 0, -publicFundBackfillDays)
	if len(movements) > 0 {
		// Most recent first: the last one is the first purchase.
		from = cashRateDay(movements[len(movements)-1].Date)
	}

	values, err := s.readPublicValues(ctx, public, from)
	if err != nil {
		return Fund{}, err
	}

	return s.repo.LinkFund(ctx, userID, assetID, public.ID, values)
}

// UnlinkFund ends a fund's link and takes back the marks it brought.
func (s *service) UnlinkFund(ctx context.Context, userID, assetID uuid.UUID) (Fund, error) {
	return s.repo.UnlinkFund(ctx, userID, assetID)
}

// ImportPublicFundValues brings every linked fund up to date with what the SFC
// published since its latest published mark: one read per published fund,
// however many owners linked it. It answers how many marks it wrote, and one
// error per published fund or linked fund that failed; the others go on.
func (s *service) ImportPublicFundValues(ctx context.Context) (int, []error) {
	if s.publicFunds == nil {
		return 0, []error{ErrPublicFundsUnavailable}
	}

	linked, err := s.repo.GetLinkedFunds(ctx)
	if err != nil {
		return 0, []error{err}
	}

	var (
		written int
		errs    []error
	)

	for _, group := range groupLinkedFunds(linked) {
		public, err := s.repo.GetPublicFund(ctx, group[0].PublicFundID)
		if err != nil {
			errs = append(errs, fmt.Errorf("public fund %s: %w", group[0].PublicFundID, err))

			continue
		}

		from := group[0].Since
		for _, l := range group[1:] {
			if l.Since.Before(from) {
				from = l.Since
			}
		}

		values, err := s.readPublicValues(ctx, public, from)
		if err != nil {
			errs = append(errs, fmt.Errorf("public fund %s: %w", public.ID, err))

			continue
		}

		for _, l := range group {
			n, err := s.repo.ImportPublicMarks(ctx, l.UserID, l.AssetID, public.ID, valuesSince(values, l.Since))
			if err != nil && !errors.Is(err, ErrFundNotFound) {
				errs = append(errs, fmt.Errorf("fund %s of user %s: %w", l.AssetID, l.UserID, err))

				continue
			}

			written += n
		}
	}

	return written, errs
}

// readPublicValues reads what a fund of the catalog published from a day on.
func (s *service) readPublicValues(ctx context.Context, public PublicFund, from time.Time) ([]PublicFundValue, error) {
	if s.publicFunds == nil {
		return nil, ErrPublicFundsUnavailable
	}

	published, err := s.publicFunds.UnitValues(ctx, public.Key(), from)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPublicFundsUnavailable, err)
	}

	return publicFundValues(published), nil
}

// preparePublicFund readies a fund that is created linked: the published fund
// must exist, the name defaults to its own, and the values it published since
// the purchase are read — the purchase's unit value among them, when the owner
// left it out.
func (s *service) preparePublicFund(ctx context.Context, in NewFundInput, now time.Time) (NewFundInput, error) {
	if in.Tracking != FundUnits || in.Currency != publicFundCurrency {
		return in, ErrFundNotLinkable
	}

	if err := validateFundDate(in.Date, now, invalidFund); err != nil {
		return in, err
	}

	public, err := s.repo.GetPublicFund(ctx, in.PublicFundID)
	if err != nil {
		return in, err
	}

	if in.CleanName() == "" {
		in.Name = truncateRunes(public.FundName, maxFundNameLen)
	}

	// Read from a few days before the purchase, whose unit value may be the
	// one published before a long weekend; only what follows it becomes marks,
	// or the fund would show a return from before the owner held it.
	day := cashRateDay(in.Date)

	values, err := s.readPublicValues(ctx, public, day.AddDate(0, 0, -publicFundOpeningLookback))
	if err != nil {
		return in, err
	}

	if in.UnitValue.IsZero() {
		v, ok := publicValueOn(values, day)
		if !ok {
			return in, invalidFund("unitValue is required: the SFC published none for %s", day.Format(time.DateOnly))
		}

		in.UnitValue = v
	}

	in.publicValues = valuesSince(values, day)

	return in, nil
}

// groupLinkedFunds groups the linked funds by the published fund they follow,
// in the order they came (GetLinkedFunds sorts by it).
func groupLinkedFunds(linked []LinkedFund) [][]LinkedFund {
	var groups [][]LinkedFund

	for _, l := range linked {
		if n := len(groups); n > 0 && groups[n-1][0].PublicFundID == l.PublicFundID {
			groups[n-1] = append(groups[n-1], l)

			continue
		}

		groups = append(groups, []LinkedFund{l})
	}

	return groups
}

// valuesSince is the values from a day on. values are oldest first.
func valuesSince(values []PublicFundValue, since time.Time) []PublicFundValue {
	since = cashRateDay(since)

	for i, v := range values {
		if !v.Date.Before(since) {
			return values[i:]
		}
	}

	return nil
}

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}

	return string([]rune(s)[:n])
}
