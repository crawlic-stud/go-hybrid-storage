package handlers

import (
	"errors"
	"fmt"
	"hybrid-storage/models"
	"hybrid-storage/utils"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func GetFileId(request *http.Request) (string, error) {
	fileId := request.PathValue("id")

	if fileId == "" {
		return fileId, errors.New("File ID is required")
	}

	return fileId, nil
}

func ReadFile(request *http.Request, filename string) ([]byte, error) {

	fileId, err := GetFileId(request)

	if err != nil {
		return nil, err
	}

	filebytes, err := os.ReadFile(filepath.Join(FILES_DIR, fileId, filename))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Not found: %v", filename))
	}

	return filebytes, nil
}

func GetFileHandler(writer http.ResponseWriter, request *http.Request) {
	metadataFile, err := ReadFile(request, METADATA_FILE)
	if err != nil {
		utils.WriteResponseStatusCode(models.Error{Detail: err.Error()}, http.StatusNotFound, writer)
		return
	}

	metadata := utils.ReadJsonData[models.FileMetadata](metadataFile)

	writer.Header().Set("Content-Type", "application/octet-stream")
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", metadata.Filename+metadata.Extension))
	http.ServeFile(writer, request, filepath.Join(FILES_DIR, metadata.FileId, "file"))
}

func GetMetadataHandler(writer http.ResponseWriter, request *http.Request) {
	metadata, err := ReadFile(request, METADATA_FILE)
	if err != nil {
		utils.WriteResponseStatusCode(models.Error{Detail: err.Error()}, http.StatusNotFound, writer)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(metadata)
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

func GetAllFilesHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	dir, err := os.Open(FILES_DIR)
	if err != nil {
		utils.WriteResponseStatusCode(models.Error{Detail: err.Error()}, http.StatusInternalServerError, writer)
		return
	}
	defer dir.Close()

	pageSize := request.URL.Query().Get("pageSize")
	// pageSize starts from 0, ends with 100
	pageSizeInt := convertToIntWithDefaultMax(pageSize, 0, 100)

	page := request.URL.Query().Get("page")
	// page starts from 1, ends never
	pageInt := convertToIntWithDefaultMax(page, 1, 0)

	// skip dirs
	for range pageInt - 1 {
		dir.ReadDir(pageSizeInt)
	}

	emptyResponse := models.PaginatedItems[models.File]{Items: []models.File{}, Page: int64(pageInt), PageSize: int64(pageSizeInt), IsNextPage: false}

	// read for page
	filesDir, err := dir.ReadDir(pageSizeInt)
	if err != nil {
		utils.WriteJsonResponse(emptyResponse, writer)
		return
	}

	var filesMetadata []models.FileMetadata
	for _, dirOrFile := range filesDir {
		if dirOrFile.IsDir() {
			metadataFile, err := os.ReadFile(filepath.Join(FILES_DIR, dirOrFile.Name(), METADATA_FILE))
			if err != nil {
				utils.WriteResponseStatusCode(models.Error{Detail: err.Error()}, http.StatusInternalServerError, writer)
				return
			}

			metadata := utils.ReadJsonData[models.FileMetadata](metadataFile)
			filesMetadata = append(filesMetadata, metadata)
		}
	}

	if filesMetadata == nil {
		utils.WriteJsonResponse(emptyResponse, writer)
		return
	}

	nextPage := false
	_, err = dir.ReadDir(pageSizeInt)
	if err == nil {
		nextPage = true
	}

	utils.WriteJsonResponse(models.PaginatedItems[models.FileMetadata]{Items: filesMetadata, Page: int64(pageInt), PageSize: int64(pageSizeInt), IsNextPage: nextPage}, writer)
}
