package utils

import (
	"hybrid-storage/models"
	"net/http"
)

func HttpHandler(function func(http.ResponseWriter, *http.Request), method string) func(http.ResponseWriter, *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			WriteResponseStatusCode(models.Error{Detail: "Method not allowed"}, http.StatusMethodNotAllowed, w)
			return
		}
		function(w, r)
	}
	return handler
}
