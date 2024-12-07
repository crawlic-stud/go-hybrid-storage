package models

type File struct {
	FileId string `json:"fileId"`
	Path   string `json:"path"`
}

type FileMetadata struct {
	FileId    string `json:"fileId"`
	Filename  string `json:"filename"`
	Extension string `json:"extension"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type FileMetadataUpdate struct {
	Filename string `json:"filename"`
}
