// Filename: internal/data/balance.go
package data

import (
	"context"
	"database/sql"
	"time"
	"github.com/spector-asael/banking/internal/validator"
)

// Balance represents the balance of a GL account.
type Balance struct {
	GLAccountID int64   `json:"gl_account_id"`
	Amount      float64 `json:"amount"`
}

// ValidateBalance checks that the balance request is valid.
func ValidateBalance(v *validator.Validator, b *Balance) {
	v.Check(b.GLAccountID > 0, "gl_account_id", "must be provided")
}

// BalanceModel wraps a DB connection pool.
type BalanceModel struct {
	DB *sql.DB
}

// GetByGLAccountID retrieves the balance for a given GL account.
func (m BalanceModel) GetByGLAccountID(glAccountID int64) (*Balance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT 
			COALESCE(SUM(debit), 0) - COALESCE(SUM(credit), 0)
		FROM ledger_entries
		WHERE gl_account_id = $1
	`

	var amount float64
	err := m.DB.QueryRowContext(ctx, query, glAccountID).Scan(&amount)
	if err != nil {
		return nil, err
	}

	return &Balance{
		GLAccountID: glAccountID,
		Amount:      amount,
	}, nil
}

// Delete removes a deposit by ledger entry ID.
// WARNING: For demonstration purposes only. In real banking systems, you usually create a reversal instead.
func (m DepositModel) DeleteByLedgerID(ledgerID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Get the journal_entry_id for the ledger entry
	var journalID int64
	err = tx.QueryRowContext(ctx,
		`SELECT journal_entry_id FROM ledger_entries WHERE id = $1`,
		ledgerID,
	).Scan(&journalID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete the ledger entry
	_, err = tx.ExecContext(ctx,
		`DELETE FROM ledger_entries WHERE id = $1`,
		ledgerID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete the corresponding journal entry
	_, err = tx.ExecContext(ctx,
		`DELETE FROM journal_entries WHERE id = $1`,
		journalID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}