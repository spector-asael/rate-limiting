// Filename: cmd/api/checkbalance.go
package handler

import (
	"net/http"
	"github.com/spector-asael/banking/internal/validator"
    "github.com/spector-asael/banking/internal/data"
)

func (a *ApplicationDependencies) checkBalanceHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	var input struct {
		GLAccountID int64 `json:"gl_account_id"`
	}

	// Decode JSON
	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}


    v := validator.New()
    data.ValidateBalance(v, &data.Balance{GLAccountID: input.GLAccountID})
    if !v.IsEmpty() {
        a.failedValidationResponse(w, r, v.Errors)
        return
    }
	// Get balance from ledger
	balance, err := a.Models.Balances.GetByGLAccountID(input.GLAccountID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Return balance
	err = a.writeJSON(
		w,
		http.StatusOK,
		envelope{"balance": balance},
		nil,
	)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}