// Filename: cmd/api/checkhistory.go
package handler

import (
	"net/http"
	"github.com/spector-asael/banking/internal/data"
	"github.com/spector-asael/banking/internal/validator"
)

func (a *ApplicationDependencies) checkHistoryHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	var input struct {
		UserID int64 `json:"user_id"`
		AccountNumber int64 `json:"account_number"`
		data.Filters 
	}

	// Decode JSON
	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if input.Page == 0 {
    input.Page = 1
	}
	if input.PageSize == 0 {
		input.PageSize = 5
	}

	input.SortSafeList = []string{
		"created_at",
		"-created_at",
		"id",
		"-id",
		"debit",
		"-debit",
		"credit",
		"-credit",
	}
	// Validate input using the model's validator
	v := validator.New()
	data.ValidateHistory(v, input.UserID)
	data.ValidateFilters(v, input.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Get ledger history
	history, metadata, err := a.Models.History.GetByUserID(input.UserID, input.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Return JSON response
	err = a.writeJSON(
		w,
		http.StatusOK,
		envelope{"history": history, "@metadata": metadata},
		nil,
	)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}