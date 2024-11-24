package handlers

import (
	"fmt"
	"hybrid-storage/models"
	"hybrid-storage/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const PERMISSIONS = 0755

func UploadFile(writer http.ResponseWriter, request *http.Request) {
	file, header, err := request.FormFile("file")
	if err != nil {
		utils.WriteResponseStatusCode(models.Error{Detail: "Error reading file"}, http.StatusBadRequest, writer)
		return
	}
	defer file.Close()

	fileId := uuid.New().String()
	path := filepath.Join("files", fileId)

	err = os.MkdirAll(path, PERMISSIONS)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
	}

	outFile, err := os.Create(filepath.Join(path, "file"))
	if err != nil {
		utils.WriteResponseStatusCode(models.Error{Detail: "Error saving file"}, http.StatusInternalServerError, writer)
		return
	}
	defer outFile.Close()

	timeNow := time.Now().UTC().Unix()
	jsonData := utils.GetJsonData(
		models.FileMetadata{
			FileId:    fileId,
			Filename:  filepath.Base(header.Filename),
			Extension: filepath.Ext(header.Filename),
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		},
	)
	err = os.WriteFile(filepath.Join(path, "metadata.json"), jsonData, PERMISSIONS)
	if err != nil {
		utils.WriteResponseStatusCode(models.Error{Detail: "Error writing metadata file"}, http.StatusInternalServerError, writer)
		return
	}

	_, err = io.Copy(outFile, file)
	if err != nil {
		utils.WriteResponseStatusCode(models.Error{Detail: "Error copying file"}, http.StatusInternalServerError, writer)
		return
	}

	utils.WriteJsonResponse(models.File{FileId: fileId, Path: path}, writer)
}
