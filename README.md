# Go Hybrid Storage

## Запуск:

```sh
go run main.go          # хранение в файловой системе
go run main.go sqlite   # хранение в SQLite
go run main.go postgres # хранение в PostgreSQL
go run main.go mongo    # хранение в MongoDB
```

## Запуск тестов:

```sh
k6 run load_tests/scripts/get_all.js
k6 run load_tests/scripts/get_random_file.js
k6 run load_tests/scripts/upload_large_file.js
k6 run load_tests/scripts/upload_one_chunk.js
```
