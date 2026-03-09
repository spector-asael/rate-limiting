// Filename: cmd/api/deposits.go
package handler

import (
	"net/http"
	"github.com/spector-asael/banking/internal/data"
	"github.com/spector-asael/banking/internal/validator"
)

func (a *ApplicationDependencies) depositHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	var input data.Deposit

	// Decode JSON into struct
	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Validate input
	v := validator.New()
	data.ValidateDeposit(v, &input)

	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert into database (creates journal + ledger entry)
	err = a.Models.Deposits.Insert(&input)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Return created deposit
	err = a.writeJSON(
		w,
		http.StatusCreated,
		envelope{"deposit": input},
		nil,
	)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}