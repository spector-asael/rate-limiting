// Filename: internal/data/deposit.go
package data

import (
	"context"
	"database/sql"
	"time"
	"github.com/spector-asael/banking/internal/validator"
)

// Deposit represents a deposit recorded in ledger_entries.
type Deposit struct {
	ID          int64     `json:"id"`             // ledger entry ID
	GLAccountID int64     `json:"gl_account_id"`  // GL account affected
	Amount      float64   `json:"amount"`         // positive = debit, negative = credit
	CreatedAt   time.Time `json:"created_at"`     // timestamp
}

// ValidateDeposit checks that the deposit is valid.
func ValidateDeposit(v *validator.Validator, d *Deposit) {
	v.Check(d.GLAccountID > 0, "gl_account_id", "must be provided")
	v.Check(d.Amount != 0, "amount", "must not be zero")
}

// DepositModel wraps a DB connection pool.
type DepositModel struct {
	DB *sql.DB
}

// Insert creates a journal entry and its corresponding ledger entry.
// Both inserts are executed inside a transaction to maintain consistency.
func (m DepositModel) Insert(d *Deposit) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// 1️⃣ Create journal entry
	var journalID int64
	journalQuery := `
		INSERT INTO journal_entries (reference_type_id, reference_id, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	// reference_type_id = 1 (assume 1 represents "deposit" in journal_reference_types)
	err = tx.QueryRowContext(ctx, journalQuery, 1, 0, "Deposit transaction").
		Scan(&journalID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2️Convert Amount into debit/credit
	var debit, credit float64
	if d.Amount > 0 {
		debit = d.Amount
		credit = 0
	} else {
		debit = 0
		credit = -d.Amount
	}

	// 3️Insert ledger entry
	ledgerQuery := `
		INSERT INTO ledger_entries (gl_account_id, journal_entry_id, debit, credit)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err = tx.QueryRowContext(ctx, ledgerQuery,
		d.GLAccountID,
		journalID,
		debit,
		credit,
	).Scan(&d.ID, &d.CreatedAt)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// UpdateAmount updates the amount of an existing ledger entry.
// For demonstration purposes only (real systems would use reversals).
func (m DepositModel) UpdateAmount(ledgerID int64, newAmount float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Convert newAmount into debit/credit
	var debit, credit float64
	if newAmount > 0 {
		debit = newAmount
		credit = 0
	} else {
		debit = 0
		credit = -newAmount
	}

	// Update ledger entry
	result, err := tx.ExecContext(ctx, `
		UPDATE ledger_entries
		SET debit = $1,
		    credit = $2,
		    updated_at = NOW()
		WHERE id = $3
	`, debit, credit, ledgerID)

	if err != nil {
		tx.Rollback()
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return sql.ErrNoRows
	}

	return tx.Commit()
}