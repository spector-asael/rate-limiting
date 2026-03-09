// Filename: cmd/api/updatedeposit.go
package handler

import (
	"errors"
	"net/http"
	"github.com/spector-asael/banking/internal/validator"
	"database/sql"

)

func (a *ApplicationDependencies) updateDepositHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	var input struct {
		LedgerID int64   `json:"ledger_id"`
		Amount   float64 `json:"amount"`
	}

	// Read JSON
	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Validate
	v := validator.New()
	v.Check(input.LedgerID > 0, "ledger_id", "must be provided")
	v.Check(input.Amount != 0, "amount", "must not be zero")

	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Update in DB
	err = a.Models.Deposits.UpdateAmount(input.LedgerID, input.Amount)
	if errors.Is(err, ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		a.failedValidationResponse(w, r, map[string]string{
			"ledger_id": "record not found",
		})
		return
	}
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Success response
	err = a.writeJSON(
		w,
		http.StatusOK,
		envelope{"message": "deposit updated successfully"},
		nil,
	)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}