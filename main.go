package main

import (
	"fmt"
	"hybrid-storage/handlers"
	fileHandlers "hybrid-storage/handlers/backends"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rs/cors"
)

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		lrw := &LoggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status code
		}

		start := time.Now()
		next.ServeHTTP(lrw, r)
		log.Printf("%d %s %s %s in %v", lrw.statusCode, r.Method, r.URL.Path, r.URL.Query().Encode(), time.Since(start))
	})
}

func determineBackendFromArgs(args []string) fileHandlers.FileServerBackend {
	var backend fileHandlers.FileServerBackend
	if len(args) > 1 {
		switch args[1] {
		case "sqlite":
			sqliteBackend, err := fileHandlers.NewSQLiteBackend("test.db")
			if err != nil {
				panic(err)
			}
			log.Println("Connected to SQLite")
			backend = sqliteBackend
		case "postgres":
			postgresBackend, err := fileHandlers.NewPostgresBackend("localhost", 5432, "postgres", "password", "postgres", "disable")
			if err != nil {
				panic(err)
			}
			log.Println("Connected to Postgres")
			backend = postgresBackend
		case "mongodb", "mongo":
			mongoDbBackend, err := fileHandlers.NewMongoDBBackend("mongodb://localhost:27017", "test")
			if err != nil {
				panic(err)
			}
			log.Println("Connected to MongoDB")
			backend = mongoDbBackend
		default:
			log.Println("Unknown backend specified, defaulting to filesystem")
			backend = fileHandlers.FileSystemBackend{}
		}
	} else {
		log.Println("No backend specified, defaulting to filesystem")
		backend = fileHandlers.FileSystemBackend{}
	}
	return backend
}

func main() {
	portServe := ":8008"
	fmt.Println("Starting server on http://localhost" + portServe)

	backend := determineBackendFromArgs(os.Args)

	// default is filesystem
	handler := http.NewServeMux()
	handler.HandleFunc("GET /", handlers.Root)

	// handlers for files
	app := handlers.App{Backend: backend, Config: handlers.AppConfig{MaxChunkSize: 5 * 1024 * 1024}}
	handler.HandleFunc("POST /files", app.UploadFileHandler)
	handler.HandleFunc("GET /files", app.GetAllFilesHandler)
	handler.HandleFunc("GET /files/{id}", app.GetFileHandler)
	handler.HandleFunc("PUT /files/{id}", app.UpdateFileHandler)
	handler.HandleFunc("DELETE /files/{id}", app.DeleteFileHandler)

	// handlers for metadata
	handler.HandleFunc("GET /files/{id}/metadata", app.GetFileMetadataHandler)

	corsConfig := cors.New(cors.Options{
		AllowedHeaders:   []string{"Origin", "Authorization", "Accept", "Content-Type"},
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "PUT"},
		AllowCredentials: true,
	})

	loggingHandler := LoggingMiddleware(handler)
	corsHandler := corsConfig.Handler(loggingHandler)

	// client := &http.Client{
	// 	Timeout: 10 * time.Second,
	// 	Transport: &http.Transport{
	// 		MaxIdleConns:        100,
	// 		MaxIdleConnsPerHost: 100,
	// 	},
	// }

	http.ListenAndServe(
		portServe,
		corsHandler,
	)
}
