# Blog API

REST API для блог-платформы на Go.

Проект включает регистрацию и авторизацию пользователей, создание постов и комментариев, отложенное логирование через горутину и хранение данных в JSON-файлах.

## Структура проекта

```
├── main.go              # Точка входа, настройка маршрутов и запуск сервера
├── models/
│   └── models.go        # Структуры данных: User, Post, Comment
├── storage/
│   └── storage.go       # Хранилище данных в JSON-файлах с мьютексом
├── handlers/
│   ├── handlers.go      # Общие хелперы и структура Handler
│   ├── auth.go          # Обработчики регистрации и входа
│   ├── posts.go         # Обработчики постов
│   ├── comments.go      # Обработчики комментариев
│   └── health.go        # Проверка состояния сервиса
├── auth/
│   └── auth.go          # Генерация и валидация JWT-токенов
├── logger/
│   └── logger.go        # Отложенное логирование через канал и горутину
├── data/                # Директория для JSON-файлов с данными
├── .env                 # Переменные окружения
├── Dockerfile           # Сборка Docker-образа
├── docker-compose.yml   # Запуск через Docker Compose
└── README.md
```

## Технологии

- Go 1.22+
- `net/http` — стандартный HTTP-сервер
- `encoding/json` — работа с JSON
- `golang.org/x/crypto/bcrypt` — хеширование паролей
- `github.com/golang-jwt/jwt/v5` — JWT-токены
- `github.com/joho/godotenv` — загрузка переменных окружения
- `sync.Mutex` — потокобезопасный доступ к хранилищу
- Горутина + канал — отложенное логирование действий

## Запуск

### Локально

```bash
# Установить зависимости
go mod tidy

# Запустить сервер
go run main.go
```

Сервер запустится на `http://localhost:8080`.

### Docker

```bash
# Собрать и запустить
docker-compose up --build

# Или в фоновом режиме
docker-compose up --build -d
```

## Переменные окружения

| Переменная    | Описание                    | По умолчанию |
|---------------|-----------------------------|--------------|
| `SERVER_HOST` | Хост сервера                | `0.0.0.0`    |
| `SERVER_PORT` | Порт сервера                | `8080`       |
| `JWT_SECRET`  | Секретный ключ для JWT      | `default-secret-key` |

## API-эндпоинты

### Состояние сервиса

```
GET /health
```

Ответ:
```json
{"status": "ok"}
```

### Регистрация

```
POST /register
Content-Type: application/json

{
  "username": "john",
  "email": "john@example.com",
  "password": "secret123"
}
```

Ответ (201):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "john",
    "email": "john@example.com"
  }
}
```

### Авторизация

```
POST /login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "secret123"
}
```

Ответ (200):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "john",
    "email": "john@example.com"
  }
}
```

### Создание поста (требуется авторизация)

```
POST /posts
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Мой первый пост",
  "content": "Содержание поста"
}
```

Ответ (201):
```json
{
  "id": 1,
  "author_id": 1,
  "title": "Мой первый пост",
  "content": "Содержание поста",
  "created_at": "2026-03-21T10:00:00Z"
}
```

### Получение всех постов

```
GET /posts
```

Ответ (200):
```json
[
  {
    "id": 1,
    "author_id": 1,
    "title": "Мой первый пост",
    "content": "Содержание поста",
    "created_at": "2026-03-21T10:00:00Z"
  }
]
```

### Получение поста по ID

```
GET /posts/{id}
```

Ответ (200):
```json
{
  "id": 1,
  "author_id": 1,
  "title": "Мой первый пост",
  "content": "Содержание поста",
  "created_at": "2026-03-21T10:00:00Z"
}
```

### Создание комментария (требуется авторизация)

```
POST /posts/{id}/comments
Authorization: Bearer <token>
Content-Type: application/json

{
  "text": "Отличный пост!"
}
```

Ответ (201):
```json
{
  "id": 1,
  "post_id": 1,
  "author_id": 1,
  "text": "Отличный пост!",
  "created_at": "2026-03-21T10:05:00Z"
}
```

### Получение комментариев к посту

```
GET /posts/{id}/comments
```

Ответ (200):
```json
[
  {
    "id": 1,
    "post_id": 1,
    "author_id": 1,
    "text": "Отличный пост!",
    "created_at": "2026-03-21T10:05:00Z"
  }
]
```

## Тестирование с curl

```bash
# Проверка состояния
curl http://localhost:8080/health

# Регистрация пользователя
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john","email":"john@example.com","password":"secret123"}'

# Вход
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123"}'

# Создание поста (подставьте свой токен)
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"title":"Мой пост","content":"Текст поста"}'

# Получение всех постов
curl http://localhost:8080/posts

# Получение поста по ID
curl http://localhost:8080/posts/1

# Создание комментария
curl -X POST http://localhost:8080/posts/1/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"text":"Отличный пост!"}'

# Получение комментариев к посту
curl http://localhost:8080/posts/1/comments
```

## Логирование

При создании поста или комментария событие отправляется в канал. Горутина-воркер читает события из канала с задержкой 1 секунду и записывает их в файл `log.txt`.

Формат записи:
```
[2026-03-21T10:00:00+03:00] user 1 created post 1
[2026-03-21T10:05:00+03:00] user 1 created comment 1
```

При остановке сервера горутина корректно завершает работу, обрабатывая оставшиеся события.

## Обработка ошибок

Все ошибки возвращаются в формате JSON:

```json
{"error": "описание ошибки"}
```

HTTP-коды:
- `200` — успех
- `201` — создано
- `400` — некорректный запрос (невалидные данные)
- `401` — не авторизован (отсутствует или невалидный токен)
- `404` — не найдено (пост/комментарий)
- `409` — конфликт (email/username уже существует)
- `500` — внутренняя ошибка сервера
