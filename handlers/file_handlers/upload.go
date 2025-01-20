package handlers

import (
	"fmt"
	"hybrid-storage/models"
	"hybrid-storage/utils"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const PERMISSIONS = 0755

// file size limit
const MB = 1024 * 1024
const MAX_CHUNK_SIZE_MB = 5
const MAX_FILE_SIZE = MAX_CHUNK_SIZE_MB * MB

const FILES_DIR = "files"
const METADATA_FILE = "metadata.json"

func saveFileToServerInChunks(writer http.ResponseWriter, request *http.Request, fileId string) {
	request.Body = http.MaxBytesReader(writer, request.Body, MAX_FILE_SIZE)

	err := request.ParseMultipartForm(MAX_FILE_SIZE)
	if err != nil {
		utils.WriteResponseStatusCode(
			models.Error{
				Detail: fmt.Sprintf("File chunk is too large, limit is %v MB", MAX_CHUNK_SIZE_MB),
			},
			http.StatusBadRequest,
			writer,
		)
		return
	}

	file, _, err := request.FormFile("file")
	if err != nil {
		utils.WriteResponseStatusCode(
			models.Error{Detail: "Error reading file"},
			http.StatusBadRequest,
			writer,
		)
		return
	}
	defer file.Close()

	chunkNum := request.FormValue("chunkNumber")
	totalChunks := request.FormValue("totalChunks")
	filenameFormValue := request.FormValue("filename")
	chunkNumInt, err := strconv.Atoi(chunkNum)
	if err != nil {
		utils.WriteResponseStatusCode(
			models.Error{Detail: "Expected int for chunk number"},
			http.StatusUnprocessableEntity,
			writer,
		)
		return
	}
	if chunkNumInt > 1 {
		fileId = request.FormValue("fileId")
	}

	path := filepath.Join(FILES_DIR, fileId)
	err = os.MkdirAll(path, PERMISSIONS)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
	}

	outFile, err := os.OpenFile(filepath.Join(path, "file"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, PERMISSIONS)
	if err != nil {
		utils.WriteResponseStatusCode(
			models.Error{Detail: "Error saving file"},
			http.StatusInternalServerError,
			writer,
		)
		return
	}
	defer outFile.Close()

	timeNow := time.Now().UTC().Unix()
	filename := filepath.Base(filenameFormValue)
	extension := filepath.Ext(filenameFormValue)
	jsonData := utils.GetJsonData(
		models.FileMetadata{
			FileId:    fileId,
			Filename:  filename[:len(filename)-len(extension)],
			Extension: extension,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		},
	)
	err = os.WriteFile(filepath.Join(path, METADATA_FILE), jsonData, PERMISSIONS)
	if err != nil {
		utils.WriteResponseStatusCode(
			models.Error{Detail: "Error writing metadata file"},
			http.StatusInternalServerError,
			writer,
		)
		return
	}

	_, err = io.Copy(outFile, file)
	if err != nil {
		utils.WriteResponseStatusCode(
			models.Error{Detail: "Error copying file"},
			http.StatusInternalServerError,
			writer,
		)
		return
	}
	log.Printf("Chunk %s of %s uploaded successfully", chunkNum, totalChunks)
	utils.WriteJsonResponse(
		models.File{FileId: fileId, Path: path},
		writer,
	)
}

func UploadFile(writer http.ResponseWriter, request *http.Request) {
	fileId := uuid.New().String()
	saveFileToServerInChunks(writer, request, fileId)
}
