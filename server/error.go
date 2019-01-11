package server

import "net/http"

func notFoundEndpoint(response http.ResponseWriter) {
	http.Error(response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func invalidMethodEndpoint(response http.ResponseWriter) {
	http.Error(response, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func ServerErrorEndpoint(response http.ResponseWriter) {
	http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
