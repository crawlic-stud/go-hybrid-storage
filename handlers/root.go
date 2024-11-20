package handlers

import (
	"net/http"
)

func Root(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, "index.html")
}
