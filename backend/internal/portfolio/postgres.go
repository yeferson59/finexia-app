package portfolio

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/gofinance/v2/money"
)

// PostgresRepository implements Repository over the shared pgx pool. Its
// methods are split by sub-area across the postgres_*.go files.
type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return new(PostgresRepository{db})
}

// currencyParam is how "no currency asked for" reaches a query: the empty
// string, which is what the COALESCE + NULLIF pattern these queries share
// reads as "use the account's preferred one".
//
// It exists because money.XXX cannot be passed straight through. XXX is the
// ISO 4217 code for "no currency", and money.Currency's driver.Valuer encodes
// it as the literal "XXX" — a value NULLIF keeps, so the fallback never fired
// and every amount came back converted at a rate of 1 and labelled XXX. The
// report then printed the generic currency sign (¤) next to the numbers.
func currencyParam(c money.Currency) string {
	if c == money.XXX {
		return ""
	}

	return c.String()
}
