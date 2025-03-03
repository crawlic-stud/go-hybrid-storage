package backends

import (
	"database/sql"
	"errors"
	"fmt"
	"hybrid-storage/models"
	"hybrid-storage/utils"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLBackend struct {
	db    *sql.DB
	query utils.Query
}

func createTables(db *sql.DB) error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS metadata (
			file_id TEXT PRIMARY KEY,
			filename TEXT NOT NULL,
			extension TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS files (
			file_id TEXT PRIMARY KEY,
			data BLOB NOT NULL,
			FOREIGN KEY (file_id) REFERENCES metadata (file_id)
		)`,
	}
	for _, tableCreateScript := range tables {
		_, err := db.Exec(tableCreateScript)
		if err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	return nil
}

func NewSQLiteBackend(dbPath string) (*SQLBackend, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = createTables(db)
	if err != nil {
		return nil, err
	}

	return &SQLBackend{db: db, query: utils.Query{Type: utils.SQLite}}, nil
}

func NewPostgresBackend(
	host string,
	port int,
	user string,
	password string,
	dbname string,
	sslMode string,
) (
	*SQLBackend,
	error,
) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host,
		port,
		user,
		password,
		dbname,
		sslMode,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = createTables(db)
	if err != nil {
		return nil, err
	}

	return &SQLBackend{db: db}, nil
}

func (b *SQLBackend) UploadFile(
	chunk utils.ChunkResult,
	fileId string,
) (
	FileServerResult,
	error,
) {
	if !chunk.IsLastChunk {
		return FileServerResult{}, &FileServerError{
			Code:   http.StatusBadRequest,
			Detail: "only full file uploads are supported",
		}
	}

	now := time.Now().Unix()

	metadata := utils.ReadJsonData[models.FileMetadata](chunk.JsonData)
	_, err := b.db.Exec(b.query.GetCachedQuery(`
		INSERT INTO metadata (file_id, filename, extension,  created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`),
		fileId,
		metadata.Filename,
		metadata.Extension,
		now,
		now,
	)
	if err != nil {
		log.Println(err.Error())
		return FileServerResult{}, errors.New("failed to insert metadata")
	}

	fileData := utils.ReadChunkBytes(chunk)
	_, err = b.db.Exec(b.query.GetCachedQuery(`
		INSERT INTO files (file_id, data)
		VALUES (?, ?)
	`),
		fileId,
		fileData,
	)
	if err != nil {
		return FileServerResult{}, errors.New("failed to insert file")
	}

	return FileServerResult{FileId: fileId}, nil
}

func handleScanErrors(errs []error) error {
	if len(errs) == 0 {
		return errors.New("empty list provided")
	}
	for _, err := range errs {
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return &FileServerError{
					Code:   http.StatusNotFound,
					Detail: "object not found",
				}
			}
			return &FileServerError{
				Code:   http.StatusInternalServerError,
				Detail: fmt.Sprintf("failed to query: %s", err.Error()),
			}
		}
	}
	return nil
}

func (b *SQLBackend) GetFile(fileId string) (GetFileResult, error) {
	metadataRow := b.db.QueryRow(b.query.GetCachedQuery(`
		SELECT file_id, filename, extension,  created_at, updated_at
		FROM metadata
		WHERE file_id = ?
	`),
		fileId,
	)
	var metadata models.FileMetadata
	metadataScanErr := metadataRow.Scan(
		&metadata.FileId,
		&metadata.Filename,
		&metadata.Extension,
		&metadata.CreatedAt,
		&metadata.UpdatedAt,
	)

	fileDataRow := b.db.QueryRow(b.query.GetCachedQuery(`
		SELECT data FROM files WHERE file_id = ?
	`),
		fileId,
	)
	var fileData []byte
	fileDataScanErr := fileDataRow.Scan(&fileData)

	err := handleScanErrors([]error{metadataScanErr, fileDataScanErr})
	if err != nil {
		return GetFileResult{}, err
	}

	return GetFileResult{File: fileData, Metadata: metadata}, nil
}

func (b *SQLBackend) GetFileMetadata(fileId string) (
	models.FileMetadata,
	error,
) {
	row := b.db.QueryRow(b.query.GetCachedQuery(`
		SELECT file_id, filename, extension, created_at, updated_at
		FROM metadata
		WHERE file_id = ?
	`),
		fileId,
	)

	var metadata models.FileMetadata
	err := row.Scan(
		&metadata.FileId,
		&metadata.Filename,
		&metadata.Extension,
		&metadata.CreatedAt,
		&metadata.UpdatedAt,
	)
	err = handleScanErrors([]error{err})
	if err != nil {
		return models.FileMetadata{}, err
	}

	return metadata, nil
}

func paginateQuery(query string, limit int, offset int) string {
	return query + fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
}

func (b *SQLBackend) GetAllFiles(page int, pageSize int) (
	PaginatedItems[models.FileMetadata],
	error,
) {
	offset := (page - 1) * pageSize

	selectQuery := `
		SELECT file_id, filename, extension, created_at, updated_at
		FROM metadata
	`
	query := paginateQuery(selectQuery, pageSize, offset)

	rows, err := b.db.Query(query)
	if err != nil {
		return PaginatedItems[models.FileMetadata]{}, &FileServerError{
			Code:   http.StatusInternalServerError,
			Detail: fmt.Sprintf("failed to query all files: %s", err.Error()),
		}
	}
	defer rows.Close()

	var files []models.FileMetadata
	for rows.Next() {
		var metadata models.FileMetadata
		err := rows.Scan(
			&metadata.FileId,
			&metadata.Filename,
			&metadata.Extension,
			&metadata.CreatedAt,
			&metadata.UpdatedAt,
		)
		if err != nil {
			return PaginatedItems[models.FileMetadata]{}, &FileServerError{
				Code:   http.StatusInternalServerError,
				Detail: fmt.Sprintf("failed to scan file metadata: %s", err.Error()),
			}
		}
		files = append(files, metadata)
	}

	futureQuery := paginateQuery(selectQuery, 1, offset+pageSize)
	futureRow, err := b.db.Query(futureQuery)
	if err != nil {
		return PaginatedItems[models.FileMetadata]{}, &FileServerError{
			Code:   http.StatusInternalServerError,
			Detail: fmt.Sprintf("failed to query next page: %s", err.Error()),
		}
	}
	defer futureRow.Close()

	result := PaginatedItems[models.FileMetadata]{
		Items:      files,
		Page:       int64(page),
		PageSize:   int64(pageSize),
		IsNextPage: futureRow.Next(),
	}
	return result, nil
}

func (b *SQLBackend) UpdateFile(
	chunk utils.ChunkResult,
	fileId string,
	data FileMetadataUpdate,
) (
	FileServerResult,
	error,
) {
	var query string
	var args []interface{}

	if chunk.FormDataChunk == nil {
		query = b.query.GetCachedQuery(`
			UPDATE metadata
			SET filename = ?, updated_at = ?
			WHERE file_id = ?
		`)
		args = []any{data.Filename, time.Now().Unix(), fileId}
		_, err := b.db.Exec(query, args...)
		if err != nil {
			return FileServerResult{}, err
		}

	} else {
		fileData := utils.ReadChunkBytes(chunk)
		query = b.query.GetCachedQuery(`
			UPDATE files
			SET data = ?
			WHERE file_id = ?
		`)
		args = []any{fileData, fileId}
		_, err := b.db.Exec(query, args...)
		if err != nil {
			return FileServerResult{}, err
		}

		metadata := utils.ReadJsonData[models.FileMetadata](chunk.JsonData)
		query = b.query.GetCachedQuery(`
			UPDATE metadata
			SET filename = ?, updated_at = ?
			WHERE file_id = ?
		`)
		args = []any{metadata.Filename, time.Now().Unix(), fileId}
		_, err = b.db.Exec(query, args...)
		if err != nil {
			return FileServerResult{}, err
		}
	}

	_, err := b.db.Exec(query, args...)
	if err != nil {
		return FileServerResult{}, err
	}

	return FileServerResult{FileId: fileId}, nil
}

func (b *SQLBackend) DeleteFile(fileId string) (bool, error) {
	_, err := b.db.Exec(b.query.GetCachedQuery(`
		DELETE FROM files
		WHERE file_id = ?
	`),
		fileId,
	)
	if err != nil {
		return false, err
	}

	_, err = b.db.Exec(b.query.GetCachedQuery(`
		DELETE FROM metadata
		WHERE file_id = ?
	`),
		fileId,
	)
	if err != nil {
		return false, err
	}

	return true, nil
}
