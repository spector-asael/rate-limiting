// File: internal/data/balance.go

package handler

import (
	"net/http"
)

func (a *ApplicationDependencies) deleteDepositHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	var input struct {
		LedgerID int64 `json:"ledger_id"`
	}

	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	err = a.Models.Deposits.DeleteByLedgerID(input.LedgerID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"message": "ledger entry deleted"}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}