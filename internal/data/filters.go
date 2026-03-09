// Filename: internal/data/filters.go
package data 

import (
	"github.com/spector-asael/banking/internal/validator"
    "strings"
)

type Filters struct {
	Page int `json:"page"`
	PageSize int `json:"page_size"`
    Sort string `json:"sort"`
    SortSafeList []string
}

type Metadata struct {
    CurrentPage int `json:"current_page,omitempty"`
    PageSize int `json:"page_size,omitempty"`
    FirstPage int `json:"first_page,omitempty"`
    LastPage int `json:"last_page,omitempty"`
    TotalRecords int `json:"total_records,omitempty"`
}
// ValidateFilters to check to see if the data provided for the filters is valid. 
// We want to make sure that the page number is greater than zero and that the page size is between 1 and 100 (inclusive).
func ValidateFilters(v *validator.Validator, f Filters) {
   v.Check(f.Page > 0, "page", "must be greater than zero")
   v.Check(f.Page <= 500, "page", "must be a maximum of 500")
   v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
   v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

   v.Check(validator.PermittedValue(f.Sort, f.SortSafeList...), "sort","invalid sort value")
}
// calculate how many records to send back
func (f Filters) limit() int {
    return f.PageSize
}

// Get the sort order
func (f Filters) sortDirection() string {
      if strings.HasPrefix(f.Sort, "-") {
          return "DESC"
      }
      return "ASC"
}

// calculate the offset so that we remember how many records have
// been sent and how many remain to be sent
func (f Filters) offset() int {
    return (f.Page - 1) * f.PageSize
}

// Calculate the metadata
func calculateMetaData(totalRecords int, currentPage int, pageSize int) Metadata {
    if totalRecords == 0 {
        return Metadata{}
    }

    return Metadata {
        CurrentPage: currentPage,
        PageSize: pageSize,
        FirstPage: 1,
        LastPage: (totalRecords + pageSize - 1) / pageSize,
        TotalRecords: totalRecords,
   }
    
}

// Implement the sorting feature
func (f Filters) sortColumn() string {
    for _, safeValue := range f.SortSafeList {
        if f.Sort == safeValue {
            return strings.TrimPrefix(f.Sort, "-")
        }
    }
   // don't allow the operation to continue
   // if case of SQL injection attack
   panic("unsafe sort parameter: " + f.Sort)
}
