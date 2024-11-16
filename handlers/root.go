package handlers

import (
	"hybrid-storage/models"
	"hybrid-storage/utils"
	"net/http"
)

func Root(writer http.ResponseWriter, request *http.Request) {
	utils.WriteJsonResponse(models.Status{Status: true}, writer)
}
