package handlers

import (
	"hybrid-storage/models"
	"hybrid-storage/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func ReplaceFileHandler(writer http.ResponseWriter, request *http.Request) {
	fileId, err := GetFileId(request)
	if err != nil {
		utils.WriteResponseStatusCode(models.Error{Detail: err.Error()}, http.StatusBadRequest, writer)
		return
	}
	saveFileToServer(writer, request, fileId)
}

func ReplaceMetadataHandler(writer http.ResponseWriter, request *http.Request) {
	fileId, err := GetFileId(request)
	if err != nil {
		utils.WriteResponseStatusCode(models.Error{Detail: err.Error()}, http.StatusBadRequest, writer)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Unable to read request body", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	data := utils.ReadJsonData[models.FileMetadataUpdate](body)

	metadataBytes, err := ReadFile(request, METADATA_FILE)
	if err != nil {
		utils.WriteResponseStatusCode(models.Error{Detail: err.Error()}, http.StatusNotFound, writer)
		return
	}

	metadata := utils.ReadJsonData[models.FileMetadata](metadataBytes)

	metadata.Filename = data.Filename
	metadata.UpdatedAt = time.Now().Unix()

	path := filepath.Join(FILES_DIR, fileId)
	jsonData := utils.GetJsonData(metadata)
	err = os.WriteFile(filepath.Join(path, METADATA_FILE), jsonData, PERMISSIONS)
	if err != nil {
		utils.WriteResponseStatusCode(models.Error{Detail: "Error writing metadata file"}, http.StatusInternalServerError, writer)
		return
	}
	utils.WriteJsonResponse(metadata, writer)
}
