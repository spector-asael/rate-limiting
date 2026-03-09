// Filename: internal/data/history.go
package data

import (
	"context"
	"database/sql"
	"time"
	"github.com/spector-asael/banking/internal/validator"
	"fmt"
)

// LedgerEntry represents a single ledger entry for a user.
type LedgerEntry struct {
	ID             int64     `json:"id"`
	GLAccountID    int64     `json:"gl_account_id"`
	JournalEntryID int64     `json:"journal_entry_id"`
	Debit          float64   `json:"debit"`
	Credit         float64   `json:"credit"`
	CreatedAt      time.Time `json:"created_at"`
	TransactionType string   `json:"transaction_type"` // "deposit" or "withdrawal"
}

// HistoryModel wraps a DB connection pool.
type HistoryModel struct {
	DB *sql.DB
}

// ValidateHistory checks that a history request is valid.
func ValidateHistory(v *validator.Validator, userID int64) {
	v.Check(userID > 0, "user_id", "must be provided")
}

// GetByUserID returns all ledger entries for a specific user
func (m HistoryModel) GetByUserID(userID int64, f Filters) ([]*LedgerEntry, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(),
			le.id,
			le.gl_account_id,
			le.journal_entry_id,
			le.debit,
			le.credit,
			le.created_at
		FROM ledger_entries le
		JOIN gl_accounts ga ON le.gl_account_id = ga.id
		JOIN accounts a ON a.gl_account_id = ga.id
		JOIN account_ownerships ao ON a.id = ao.account_id
		JOIN customers c ON ao.customer_id = c.id
		WHERE c.person_id = $1
		ORDER BY %s %s, le.id ASC
		LIMIT $2 OFFSET $3`,
		f.sortColumn(),
		f.sortDirection(),
	)

	rows, err := m.DB.QueryContext(ctx, query, userID, f.limit(), f.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	var entries []*LedgerEntry

	for rows.Next() {
		var le LedgerEntry

		err := rows.Scan(
			&totalRecords,
			&le.ID,
			&le.GLAccountID,
			&le.JournalEntryID,
			&le.Debit,
			&le.Credit,
			&le.CreatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		switch {
		case le.Debit > 0:
			le.TransactionType = "deposit"
		case le.Credit > 0:
			le.TransactionType = "withdrawal"
		default:
			le.TransactionType = "unknown"
		}

		entries = append(entries, &le)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Calculate metadata AFTER loop
	metadata := calculateMetaData(totalRecords, f.Page, f.PageSize)

	return entries, metadata, nil
}