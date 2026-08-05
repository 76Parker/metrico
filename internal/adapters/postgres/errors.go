package postgres

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var retrayablePgErrors = map[string]struct{}{
	pgerrcode.ConnectionException:                           {},
	pgerrcode.ConnectionDoesNotExist:                        {},
	pgerrcode.ConnectionFailure:                             {},
	pgerrcode.SQLClientUnableToEstablishSQLConnection:       {},
	pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection: {},
	pgerrcode.TransactionResolutionUnknown:                  {},
	pgerrcode.ProtocolViolation:                             {},
}

func isRetryable(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	_, ok := retrayablePgErrors[pgErr.Code]
	return ok
}
