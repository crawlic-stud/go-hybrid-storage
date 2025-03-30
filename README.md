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
