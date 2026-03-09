package handler

import (
  "net/http"
  "github.com/julienschmidt/httprouter"
)

func (a *ApplicationDependencies)Routes() http.Handler  {

   // setup a new router
   router := httprouter.New()
   router.NotFound = http.HandlerFunc(a.notFoundResponse)
   router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)
   // setup routes
   router.HandlerFunc(http.MethodGet, "/v1/balance", a.checkBalanceHandler)
   router.HandlerFunc(http.MethodPost, "/v1/deposit", a.depositHandler)
   router.HandlerFunc(http.MethodPost, "/v1/history", a.checkHistoryHandler)
   router.HandlerFunc(http.MethodDelete, "/v1/delete", a.deleteDepositHandler)
   router.HandlerFunc(http.MethodPatch, "/v1/update", a.updateDepositHandler)
   router.HandlerFunc(http.MethodPost, "/v1/transfer", TransferHandler)

   loggingMiddleware := a.loggingMiddleware(router)
   panicMiddleware := a.recoverPanic(loggingMiddleware)
   return panicMiddleware     
  
}
