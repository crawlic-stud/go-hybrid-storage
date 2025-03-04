package backends

import (
	"context"
	"errors"
	"fmt"
	"hybrid-storage/models"
	"hybrid-storage/utils"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBBackend struct {
	client   *mongo.Client
	db       *mongo.Database
	metadata *mongo.Collection
	files    *mongo.Collection
}

type BSONFileChunk struct {
	FileId string `bson:"fileId"`
	Chunk  int    `bson:"chunk"`
	Data   []byte `bson:"data"`
}

func NewMongoDBBackend(
	uri string,
	dbName string,
) (*MongoDBBackend, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	db := client.Database(dbName)
	db.Drop(context.Background())
	metadataCollection := db.Collection("metadata")
	filesCollection := db.Collection("file_chunks")

	_, err = filesCollection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "fileId", Value: 1}, {Key: "chunk", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return &MongoDBBackend{
		client:   client,
		db:       db,
		metadata: metadataCollection,
		files:    filesCollection,
	}, nil
}

func (b *MongoDBBackend) UploadFile(
	chunk utils.ChunkResult,
	fileId string,
) (
	FileServerResult,
	error,
) {
	now := time.Now().Unix()

	if chunk.ChunkNumber == 1 {
		metadata := utils.ReadJsonData[models.FileMetadata](chunk.JsonData)
		_, err := b.metadata.InsertOne(context.Background(), bson.M{
			"fileId":    fileId,
			"filename":  metadata.Filename,
			"extension": metadata.Extension,
			"createdAt": now,
			"updatedAt": now,
		})
		if err != nil {
			log.Println(err.Error())
			return FileServerResult{}, errors.New("failed to insert metadata")
		}
	} else {
		fileId = chunk.FileId
	}

	_, err := b.files.InsertOne(context.Background(), BSONFileChunk{
		FileId: fileId,
		Chunk:  chunk.ChunkNumber,
		Data:   utils.ReadChunkBytes(chunk),
	})
	if err != nil {
		log.Println(err.Error())
		return FileServerResult{}, errors.New("failed to insert file chunk")
	}

	return FileServerResult{FileId: fileId}, nil
}

func (b *MongoDBBackend) GetFile(fileId string) (GetFileResult, error) {
	var metadata models.FileMetadata
	err := b.metadata.FindOne(context.Background(), bson.M{"fileId": fileId}).
		Decode(&metadata)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return GetFileResult{}, &FileServerError{
				Code:   http.StatusNotFound,
				Detail: "metadata not found",
			}
		}
		return GetFileResult{}, fmt.Errorf("failed to query metadata: %w", err)
	}

	cursor, err := b.files.Find(
		context.Background(),
		bson.M{"fileId": fileId},
		options.Find().SetSort(bson.M{"chunk": 1}),
	)
	if err != nil {
		return GetFileResult{}, fmt.Errorf("failed to query file chunks: %w", err)
	}
	defer cursor.Close(context.Background())

	var fileData []byte
	for cursor.Next(context.Background()) {
		var chunk BSONFileChunk
		err := cursor.Decode(&chunk)
		if err != nil {
			return GetFileResult{}, fmt.Errorf("failed to decode file chunk: %w", err)
		}
		fileData = append(fileData, chunk.Data...)
	}

	if err := cursor.Err(); err != nil {
		return GetFileResult{}, fmt.Errorf("cursor error: %w", err)
	}

	if len(fileData) == 0 {
		return GetFileResult{}, &FileServerError{
			Code:   http.StatusNotFound,
			Detail: "file data not found",
		}
	}

	return GetFileResult{File: fileData, Metadata: metadata}, nil
}

func (b *MongoDBBackend) GetFileMetadata(fileId string) (
	models.FileMetadata,
	error,
) {
	var metadata models.FileMetadata
	err := b.metadata.FindOne(context.Background(), bson.M{"fileId": fileId}).
		Decode(&metadata)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.FileMetadata{}, &FileServerError{
				Code:   http.StatusNotFound,
				Detail: "metadata not found",
			}
		}
		return models.FileMetadata{}, fmt.Errorf("failed to query metadata: %w", err)
	}
	return metadata, nil
}

func (b *MongoDBBackend) GetAllFiles(page int, pageSize int) (
	PaginatedItems[models.FileMetadata],
	error,
) {
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	cursor, err := b.metadata.Find(
		context.Background(),
		bson.M{},
		options.Find().SetSkip(skip).SetLimit(limit),
	)
	if err != nil {
		return PaginatedItems[models.FileMetadata]{}, &FileServerError{
			Code:   http.StatusInternalServerError,
			Detail: fmt.Sprintf("failed to query all files: %s", err.Error()),
		}
	}
	defer cursor.Close(context.Background())

	var files []models.FileMetadata
	for cursor.Next(context.Background()) {
		var metadata models.FileMetadata
		if err := cursor.Decode(&metadata); err != nil {
			return PaginatedItems[models.FileMetadata]{}, &FileServerError{
				Code:   http.StatusInternalServerError,
				Detail: fmt.Sprintf("failed to scan file metadata: %s", err.Error()),
			}
		}
		files = append(files, metadata)
	}

	if err := cursor.Err(); err != nil {
		return PaginatedItems[models.FileMetadata]{}, &FileServerError{
			Code:   http.StatusInternalServerError,
			Detail: fmt.Sprintf("cursor error: %s", err.Error()),
		}
	}

	count, err := b.metadata.CountDocuments(
		context.Background(),
		bson.M{},
		options.Count().SetSkip(skip+limit).SetLimit(1),
	)
	if err != nil {
		return PaginatedItems[models.FileMetadata]{}, &FileServerError{
			Code:   http.StatusInternalServerError,
			Detail: fmt.Sprintf("failed to query next page: %s", err.Error()),
		}
	}

	result := PaginatedItems[models.FileMetadata]{
		Items:      files,
		Page:       int64(page),
		PageSize:   int64(pageSize),
		IsNextPage: count > 0,
	}
	return result, nil
}

func (b *MongoDBBackend) UpdateFile(
	chunk utils.ChunkResult,
	fileId string,
	data FileMetadataUpdate,
) (
	FileServerResult,
	error,
) {
	// update only metadata
	if chunk.FormDataChunk == nil {
		update := bson.M{
			"$set": bson.M{
				"filename":   data.Filename,
				"updated_at": time.Now().Unix(),
			},
		}
		_, err := b.metadata.UpdateOne(
			context.Background(),
			bson.M{"fileId": fileId},
			update,
		)
		if err != nil {
			return FileServerResult{}, fmt.Errorf("failed to update metadata: %w", err)
		}
	} else { // else delete old file and upload new with same fileId
		if chunk.ChunkNumber == 1 {
			b.DeleteFile(fileId)
		}
		b.UploadFile(chunk, fileId)
	}

	return FileServerResult{FileId: fileId}, nil
}

func (b *MongoDBBackend) DeleteFile(fileId string) (bool, error) {
	_, err := b.files.DeleteMany(context.Background(), bson.M{"fileId": fileId})
	if err != nil {
		return false, fmt.Errorf("failed to delete file chunks: %w", err)
	}

	deleteResult, err := b.metadata.DeleteOne(context.Background(), bson.M{"fileId": fileId})
	if err != nil {
		return false, fmt.Errorf("failed to delete metadata: %w", err)
	}

	if deleteResult.DeletedCount == 0 {
		return false, &FileServerError{
			Code:   http.StatusNotFound,
			Detail: "file not found",
		}
	}

	return true, nil
}

func (b *MongoDBBackend) Close() error {
	return b.client.Disconnect(context.Background())
}
