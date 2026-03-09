package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"errors"
	"net/http"
	"strings"
	"database/sql"
)

func TransferHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		FromAccountID int64   `json:"from_account_id"`
		ToAccountID   int64   `json:"to_account_id"`
		Amount        float64 `json:"amount"`
	}

	// ❌ Manual JSON parsing (no DI)
	err := readJSON(w, r, &input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ❌ Manual validation
	if input.FromAccountID <= 0 || input.ToAccountID <= 0 {
		http.Error(w, "account IDs must be provided", http.StatusBadRequest)
		return
	}

	if input.Amount <= 0 {
		http.Error(w, "amount must be greater than 0", http.StatusBadRequest)
		return
	}

	// ❌ Open DB inside handler (very bad practice)
	db, err := sql.Open("postgres", "postgres://banking:banking@localhost/banking?sslmode=disable")
	if err != nil {
		http.Error(w, "database connection error", 500)
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "could not start transaction", 500)
		fmt.Println(err)
		return
	}

	var fromGL, toGL int64

	// Get sender GL
	err = tx.QueryRow(
		`SELECT gl_account_id FROM accounts WHERE id=$1`,
		input.FromAccountID,
	).Scan(&fromGL)
	if err != nil {
		tx.Rollback()
		http.Error(w, "sender account not found", 400)
		return
	}

	// Get receiver GL
	err = tx.QueryRow(
		`SELECT gl_account_id FROM accounts WHERE id=$1`,
		input.ToAccountID,
	).Scan(&toGL)
	if err != nil {
		tx.Rollback()
		http.Error(w, "receiver account not found", 400)
		return
	}

	// Create journal entry
	var journalID int64
	err = tx.QueryRow(`
		INSERT INTO journal_entries (reference_type_id, reference_id, description)
		VALUES (1, $1, 'Transfer')
		RETURNING id
	`, input.FromAccountID).Scan(&journalID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "failed to create journal entry", 500)
		return
	}

	// Sender loses money (credit)
	_, err = tx.Exec(`
		INSERT INTO ledger_entries (gl_account_id, journal_entry_id, credit)
		VALUES ($1, $2, $3)
	`, fromGL, journalID, input.Amount)
	if err != nil {
		tx.Rollback()
		http.Error(w, "failed to create sender ledger entry", 500)
		return
	}

	// Receiver gains money (debit)
	_, err = tx.Exec(`
		INSERT INTO ledger_entries (gl_account_id, journal_entry_id, debit)
		VALUES ($1, $2, $3)
	`, toGL, journalID, input.Amount)
	if err != nil {
		tx.Rollback()
		http.Error(w, "failed to create receiver ledger entry", 500)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "failed to commit transaction", 500)
		return
	}

	// ✅ Respond to user
	response := envelope{
		"message": "transfer completed successfully",
		"from_account_id": input.FromAccountID,
		"to_account_id": input.ToAccountID,
		"amount": input.Amount,
	}

	writeJSON(w, http.StatusOK, response, nil)
}

func writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {

	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(js)
	return err
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	maxBytes := 256_000
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {

		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {

		case errors.As(err, &syntaxError):
			return fmt.Errorf("badly formed JSON (at position %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("badly formed JSON")

		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf("incorrect JSON type for field %q", unmarshalTypeError.Field)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("unknown field %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not exceed %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// Check for multiple JSON objects
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must contain only a single JSON value")
	}

	return nil
}