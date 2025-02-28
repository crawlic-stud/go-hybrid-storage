package utils

import (
	"errors"
	"fmt"
	"hybrid-storage/models"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

type ChunkResult struct {
	FormDataChunk multipart.File
	ChunkNumber   int
	FileId        string
	IsLastChunk   bool
	JsonData      []byte
}

func ReadFileInChunks(writer http.ResponseWriter, request *http.Request, fileId string, maxFileSize int64, maxChunkSizeMb int) (ChunkResult, error) {
	request.Body = http.MaxBytesReader(writer, request.Body, maxFileSize)

	err := request.ParseMultipartForm(maxFileSize)
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		return ChunkResult{}, fmt.Errorf("file chunk is too large, limit is %v MB", maxChunkSizeMb)
	}

	fileChunk, _, err := request.FormFile("file")
	if err != nil {
		return ChunkResult{}, errors.New("error reading file")
	}
	defer fileChunk.Close()

	chunkNum := request.FormValue("chunkNumber")
	totalChunks := request.FormValue("totalChunks")
	filenameFormValue := request.FormValue("filename")
	chunkNumInt, err := strconv.Atoi(chunkNum)
	if err != nil {
		return ChunkResult{}, errors.New("expected int for chunk number")
	}
	if chunkNumInt > 1 {
		fileId = request.FormValue("fileId")
	}
	log.Printf("Chunk %s/%s for file %s uploaded successfully", chunkNum, totalChunks, fileId)

	timeNow := time.Now().UTC().Unix()
	filename := filepath.Base(filenameFormValue)
	extension := filepath.Ext(filenameFormValue)
	jsonData := GetJsonData(
		models.FileMetadata{
			FileId:    fileId,
			Filename:  filename[:len(filename)-len(extension)],
			Extension: extension,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		},
	)

	return ChunkResult{
		FormDataChunk: fileChunk,
		ChunkNumber:   chunkNumInt,
		IsLastChunk:   chunkNum == totalChunks,
		FileId:        fileId,
		JsonData:      jsonData,
	}, nil
}

func ReadChunkBytes(chunk ChunkResult) []byte {
	bytes := make([]byte, 0)
	for {
		currentBytes := make([]byte, 1024)
		_, err := chunk.FormDataChunk.Read(currentBytes)
		if err != nil {
			break
		}
		bytes = append(bytes, currentBytes...)
	}
	return bytes
}
