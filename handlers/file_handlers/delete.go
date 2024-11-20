package handlers

import (
	"fmt"
	"hybrid-storage/models"
	"hybrid-storage/utils"
	"net/http"
	"os"
	"path/filepath"
)

func DeleteFileHandler(writer http.ResponseWriter, request *http.Request) {

	fmt.Printf("DeleteFileHandler method=%s\n", request.Method)

	fileId, err := GetFileId(request)

	if err != nil {
		utils.WriteResponseStatusCode(models.Error{Detail: err.Error()}, http.StatusBadRequest, writer)
		return
	}

	err = os.RemoveAll(filepath.Join("files", fileId))

	if err != nil {
		fmt.Println(err.Error())
		utils.WriteResponseStatusCode(models.Error{Detail: "File not found"}, http.StatusNotFound, writer)
		return
	}

	utils.WriteJsonResponse(models.Status{Status: true}, writer)
}
