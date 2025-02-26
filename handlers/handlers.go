package handlers

import (
	"fmt"
	"hybrid-storage/handlers/backends"
	"hybrid-storage/models"
	"hybrid-storage/utils"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type AppConfig struct {
	MaxFileSize    int64
	MaxChunkSizeMb int
}

type App struct {
	Backend backends.FileServerBackend
	Config  AppConfig
}

func handleBackendError(writer http.ResponseWriter, err error) {
	backendErr, ok := err.(*backends.FileServerError)
	if ok {
		utils.WriteResponseStatusCode(models.Error{Detail: backendErr.Detail}, backendErr.Code, writer)
	} else {
		utils.WriteResponseStatusCode(models.Error{Detail: err.Error()}, http.StatusInternalServerError, writer)
	}
}

func (app *App) UploadFileHandler(writer http.ResponseWriter, request *http.Request) {
	fileId := uuid.New().String()
	chunk, err := utils.ReadFileInChunks(
		writer,
		request,
		fileId,
		app.Config.MaxFileSize,
		app.Config.MaxChunkSizeMb,
	)
	if err != nil {
		handleBackendError(writer, err)
		return
	}

	result, err := app.Backend.UploadFile(chunk, fileId)
	if err != nil {
		handleBackendError(writer, err)
		return
	}
	utils.WriteJsonResponse(result, writer)
}

func (app *App) GetFileHandler(writer http.ResponseWriter, request *http.Request) {
	fileId, err := utils.GetFileId(request)
	if err != nil {
		handleBackendError(writer, err)
		return
	}
	result, err := app.Backend.GetFile(fileId)
	if err != nil {
		handleBackendError(writer, err)
		return
	}
	writer.Header().Set("Content-Type", "application/octet-stream")
	writer.Header().Set(
		"Content-Disposition",
		fmt.Sprintf(
			"attachment; filename=%q",
			result.Metadata.Filename+result.Metadata.Extension,
		),
	)
	writer.Write(result.File)
}

func (app *App) GetFileMetadataHandler(writer http.ResponseWriter, request *http.Request) {
	fileId, err := utils.GetFileId(request)
	if err != nil {
		handleBackendError(writer, err)
		return
	}
	result, err := app.Backend.GetFileMetadata(fileId)
	if err != nil {
		handleBackendError(writer, err)
		return
	}
	utils.WriteJsonResponse(result, writer)
}

func convertToIntWithDefaultMax(value string, defaultValue int, max int) int {
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	if intValue > max && max != 0 {
		return max
	}
	return intValue
}

func (app *App) GetAllFilesHandler(writer http.ResponseWriter, request *http.Request) {
	page := request.URL.Query().Get("page")
	pageInt := convertToIntWithDefaultMax(page, 1, 0)
	pageSize := request.URL.Query().Get("pageSize")
	pageSizeInt := convertToIntWithDefaultMax(pageSize, 0, 100)
	result, err := app.Backend.GetAllFiles(pageInt, pageSizeInt)
	if err != nil {
		handleBackendError(writer, err)
		return
	}
	utils.WriteJsonResponse(result, writer)
}

func (app *App) UpdateFileHandler(writer http.ResponseWriter, request *http.Request) {
	fileId, err := utils.GetFileId(request)
	if err != nil {
		handleBackendError(writer, err)
		return
	}

	chunk := utils.ChunkResult{IsLastChunk: true}
	data := backends.FileMetadataUpdate{Filename: ""}
	if request.Header.Get("Content-Type") == "application/json" {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, "Unable to read request body", http.StatusBadRequest)
			return
		}
		defer request.Body.Close()
		data = utils.ReadJsonData[backends.FileMetadataUpdate](body)
	} else {
		chunk, err = utils.ReadFileInChunks(
			writer,
			request,
			fileId,
			app.Config.MaxFileSize,
			app.Config.MaxChunkSizeMb,
		)
		if err != nil {
			handleBackendError(writer, err)
			return
		}
	}
	result, err := app.Backend.UpdateFile(chunk, fileId, data)
	if err != nil {
		handleBackendError(writer, err)
		return
	}
	utils.WriteJsonResponse(result, writer)
}

func (app *App) DeleteFileHandler(writer http.ResponseWriter, request *http.Request) {
	fileId, err := utils.GetFileId(request)
	if err != nil {
		handleBackendError(writer, err)
		return
	}
	status, err := app.Backend.DeleteFile(fileId)
	if err != nil {
		handleBackendError(writer, err)
		return
	}
	utils.WriteJsonResponse(models.Status{Status: status}, writer)
}
