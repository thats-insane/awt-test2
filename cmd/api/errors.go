package main

import (
	"fmt"
	"net/http"
)

func (a *appDependencies) logError(r *http.Request, err error) {
	method := r.Method
	uri := r.URL.RequestURI()
	a.logger.Error(err.Error(), "method", method, "uri", uri)
}

func (a *appDependencies) errResponseJSON(w http.ResponseWriter, r *http.Request, status int, message any) {
	errData := envelope{
		"error": message,
	}
	err := a.writeJSON(w, status, errData, nil)
	if err != nil {
		a.logError(r, err)
		w.WriteHeader(500)
	}
}

func (a *appDependencies) serverErrResponse(w http.ResponseWriter, r *http.Request, err error) {
	a.logError(r, err)
	message := "the server encountered a problem and could not process your request"
	a.errResponseJSON(w, r, http.StatusInternalServerError, message)
}

func (a *appDependencies) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	a.errResponseJSON(w, r, http.StatusNotFound, message)
}

func (a *appDependencies) notAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	a.errResponseJSON(w, r, http.StatusMethodNotAllowed, message)
}

func (a *appDependencies) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	a.errResponseJSON(w, r, http.StatusBadRequest, err.Error())
}

func (a *appDependencies) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	a.errResponseJSON(w, r, http.StatusUnprocessableEntity, errors)
}
