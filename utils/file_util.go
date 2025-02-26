package utils

import (
	"fmt"
	"net/http"
)

func GetFileId(request *http.Request) (string, error) {
	fileId := request.PathValue("id")

	if fileId == "" {
		return fileId, fmt.Errorf("%s", "File ID is required")
	}

	return fileId, nil
}
