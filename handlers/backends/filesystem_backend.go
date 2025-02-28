package backends

import (
	"fmt"
	"hybrid-storage/models"
	"hybrid-storage/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type FileSystemBackend struct{}

const PERMISSIONS = 0755
const FILES_DIR = "files"
const METADATA_FILE = "metadata.json"
const FILE_NAME = "file"

func (fsb FileSystemBackend) UploadFile(chunk utils.ChunkResult, fileId string) (FileServerResult, error) {
	path := filepath.Join(FILES_DIR, chunk.FileId)
	err := os.MkdirAll(path, PERMISSIONS)
	if err != nil {
		return FileServerResult{}, &FileServerError{
			Code:   http.StatusInternalServerError,
			Detail: "Error creating file directory",
		}
	}
	outFile, err := os.OpenFile(filepath.Join(path, FILE_NAME), os.O_CREATE|os.O_WRONLY|os.O_APPEND, PERMISSIONS)
	if err != nil {
		return FileServerResult{}, &FileServerError{
			Code:   http.StatusInternalServerError,
			Detail: "Error saving file",
		}
	}
	defer outFile.Close()

	err = os.WriteFile(filepath.Join(path, METADATA_FILE), chunk.JsonData, PERMISSIONS)
	if err != nil {
		return FileServerResult{}, &FileServerError{
			Code:   http.StatusInternalServerError,
			Detail: "Error writing metadata file",
		}
	}

	_, err = io.Copy(outFile, chunk.FormDataChunk)
	if err != nil {
		return FileServerResult{}, &FileServerError{
			Code:   http.StatusInternalServerError,
			Detail: "Error copying file",
		}
	}
	return FileServerResult{FileId: fileId}, nil
}

func (fsb FileSystemBackend) GetFile(fileId string) (GetFileResult, error) {
	filebytes, err := os.ReadFile(filepath.Join(FILES_DIR, fileId, FILE_NAME))
	if err != nil {
		return GetFileResult{}, &FileServerError{
			Code:   http.StatusNotFound,
			Detail: fmt.Sprintf("%s: %s", "File not found", fileId),
		}
	}

	metadataFile, err := os.ReadFile(filepath.Join(FILES_DIR, fileId, METADATA_FILE))
	if err != nil {
		return GetFileResult{}, &FileServerError{
			Code:   http.StatusNotFound,
			Detail: err.Error(),
		}
	}
	metadata := utils.ReadJsonData[models.FileMetadata](metadataFile)
	return GetFileResult{File: filebytes, Metadata: metadata}, nil
}

func (fsb FileSystemBackend) GetFileMetadata(fileId string) (models.FileMetadata, error) {
	metadataFile, err := os.ReadFile(filepath.Join(FILES_DIR, fileId, METADATA_FILE))
	if err != nil {
		return models.FileMetadata{}, &FileServerError{
			Code:   http.StatusNotFound,
			Detail: err.Error(),
		}
	}
	return utils.ReadJsonData[models.FileMetadata](metadataFile), nil
}

func (fsb FileSystemBackend) GetAllFiles(page int, pageSize int) (PaginatedItems[models.FileMetadata], error) {
	dir, err := os.Open(FILES_DIR)
	if err != nil {
		return PaginatedItems[models.FileMetadata]{}, err
	}
	defer dir.Close()

	// skip dirs
	for range page - 1 {
		dir.ReadDir(pageSize)
	}

	// read for page
	filesDir, err := dir.ReadDir(pageSize)
	if err != nil {
		return PaginatedItems[models.FileMetadata]{}, err
	}

	var filesMetadata []models.FileMetadata
	for _, dirOrFile := range filesDir {
		if dirOrFile.IsDir() {
			metadataFile, err := os.ReadFile(filepath.Join(FILES_DIR, dirOrFile.Name(), METADATA_FILE))
			if err != nil {
				return PaginatedItems[models.FileMetadata]{}, err
			}

			metadata := utils.ReadJsonData[models.FileMetadata](metadataFile)
			filesMetadata = append(filesMetadata, metadata)
		}
	}

	if filesMetadata == nil {
		return PaginatedItems[models.FileMetadata]{}, nil
	}

	nextPage := false
	_, err = dir.ReadDir(pageSize)
	if err == nil {
		nextPage = true
	}

	return PaginatedItems[models.FileMetadata]{
		Items:      filesMetadata,
		Page:       int64(page),
		PageSize:   int64(pageSize),
		IsNextPage: nextPage,
	}, nil
}

func (fsb FileSystemBackend) UpdateFile(chunk utils.ChunkResult, fileId string, metadataUpdate FileMetadataUpdate) (FileServerResult, error) {
	result, err := fsb.UploadFile(chunk, fileId)
	if err != nil {
		return FileServerResult{}, err
	}

	if chunk.IsLastChunk && (metadataUpdate.Filename != "") {
		metadataFile, err := os.ReadFile(filepath.Join(FILES_DIR, fileId, METADATA_FILE))
		if err != nil {
			return FileServerResult{}, &FileServerError{
				Code:   http.StatusInternalServerError,
				Detail: err.Error(),
			}
		}
		metadata := utils.ReadJsonData[models.FileMetadata](metadataFile)
		metadata.Filename = metadataUpdate.Filename
		metadata.UpdatedAt = time.Now().Unix()
		path := filepath.Join(FILES_DIR, fileId)
		jsonData := utils.GetJsonData(metadata)
		err = os.WriteFile(filepath.Join(path, METADATA_FILE), jsonData, PERMISSIONS)
		if err != nil {
			return FileServerResult{}, &FileServerError{
				Code:   http.StatusInternalServerError,
				Detail: "Error writing metadata file",
			}
		}
	}
	return result, nil
}

func (fsb FileSystemBackend) DeleteFile(fileId string) (bool, error) {
	err := os.RemoveAll(filepath.Join(FILES_DIR, fileId))
	if err != nil {
		return false, &FileServerError{
			Code:   http.StatusInternalServerError,
			Detail: "Error deleting file",
		}
	}
	return true, nil
}
