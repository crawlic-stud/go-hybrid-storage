# Go Hybrid Storage

## Запуск:

```sh
go run main.go
```

## Запуск тестов:

```sh
k6 run load_tests/scripts/get_all.js
k6 run load_tests/scripts/get_random_file.js
k6 run load_tests/scripts/upload_large_file.js
k6 run load_tests/scripts/upload_small_file.js
```
