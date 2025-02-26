package backends

import (
	"hybrid-storage/models"
	"hybrid-storage/utils"
)

type FileServerResult struct {
	FileId string `json:"fileId"`
}

type FileServerError struct {
	Code   int    `json:"code"`
	Detail string `json:"detail"`
}

func (e *FileServerError) Error() string {
	return e.Detail
}

type FileMetadataUpdate struct {
	Filename string `json:"filename"`
}

type PaginatedItems[T any] struct {
	Items      []T   `json:"items"`
	Page       int64 `json:"page"`
	PageSize   int64 `json:"pageSize"`
	IsNextPage bool  `json:"isNextPage"`
}

type GetFileResult struct {
	File     []byte
	Metadata models.FileMetadata
}

type FileServerBackend interface {
	UploadFile(chunk utils.ChunkResult, fileId string) (FileServerResult, error)
	UpdateFile(chunk utils.ChunkResult, fileId string, metadataUpdate FileMetadataUpdate) (FileServerResult, error)
	GetFile(fileId string) (GetFileResult, error)
	GetFileMetadata(fileId string) (models.FileMetadata, error)
	GetAllFiles(page int, pageSize int) (PaginatedItems[models.FileMetadata], error)
	DeleteFile(fileId string) (bool, error)
}
