package handler 

import (
	"net/http"
	"time"
)

func shutdownTestHandler(w http.ResponseWriter, r *http.Request) {
    time.Sleep(8 * time.Second)
	writeJSON(w, http.StatusOK, envelope{"message": "server is shutting down"}, nil)
}