# Go Hybrid Storage

Выпускная квалификационная работа Кузнецова Н. С. - НИУ ИТМО 2025

## Запуск бэкенда:

```sh
go build .                  # сборка проекта
./hybrid-storage            # хранение в файловой системе
./hybrid-storage sqlite     # хранение в SQLite
./hybrid-storage postgres   # хранение в PostgreSQL
./hybrid-storage mongo      # хранение в MongoDB
```

## Запуск фронтенда:

Перейти по `http://localhost:8008`

## Запуск всех тестов:

```sh
make test-all
```

## Запуск конкретного теста

```sh
make test=... back=...
```

> Тесты: `upload_large_chunk`, `upload_small_chunk`, `get_random_file`

> Бэкенды: `fs`, `sqlite`, `postgres`, `mongo`

## Структура бэкендов

```sh
├── handlers
│   ├── backends                    # модуль бэкендов, реализующих операции с файлами
│   │   ├── filesystem_backend.go   # реализация бэкенда файловой системы
│   │   ├── interface.go            # общий интерфейс для всех бэкендов
│   │   ├── mongodb_backend.go      # реализация бэкенда MongoDB
│   │   └── sql_backend.go          # реализация бэкенда SQL
│   ├── handlers.go                 # хендлеры для взаимодействия API с конкретным бэкендом
│   └── root.go                     # основной хендлер - для фронтенда
```
