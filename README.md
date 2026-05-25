# TaskFlow

Приложение для совместного управления задачами.

## Возможности

- Регистрация и вход через JWT
- Создание проектов и приглашение участников (роли: owner / admin / member)
- Задачи: создание, назначение, статусы, приоритеты, фильтрация, сортировка, пагинация
- Комментарии к задачам
- Статистика по проектам и пользователям 
- Уведомления через отдельный gRPC-сервис и пул воркеров

## Структура проекта

```
taskflow/
├── cmd/app/                 # точка входа 
├── internal/
│   ├── handler/             # HTTP-обработчики и роутер
│   ├── service/             # бизнес-логика
│   ├── repository/          # доступ к БД 
│   ├── middleware/          # JWT, логирование, recovery, rate-limit, request-id
│   ├── model/  dto/         # доменные сущности и payloadы
│   ├── worker/              # пул воркеров и шина событий
│   ├── cache/               # Redis
│   ├── grpc/{client,server} # gRPC-клиент и сервер уведомлений
│   ├── config/ logger/ auth/ utils/
├── migrations/              # SQL-миграции
├── proto/                   # proto и сгенерированный пакет
├── tests/                   # юнит-тесты + моки
└── docker-compose.yml  Dockerfile  Makefile  .env.example
```


## Старт

```bash
cp .env.example .env
make docker-up       
make migrate-up      
```

Остановить:

```bash
make docker-down
```

## Миграции

```bash
make migrate-up                              # применить все
make migrate-down                            # откатить последнюю
make migrate-create NAME=add_labels          # создать новую
```

## API

Базовый путь: /api/v1. Все маршруты, кроме /auth/register и /auth/login,
требуют заголовок Authorization: Bearer <JWT>.

### Auth

```http
POST /api/v1/auth/register   { "email","password","name" }
POST /api/v1/auth/login      { "email","password" }
GET  /api/v1/auth/me
```

### Projects

```http
POST   /api/v1/projects
GET    /api/v1/projects?page=1&limit=20
GET    /api/v1/projects/:id
PUT    /api/v1/projects/:id
DELETE /api/v1/projects/:id

POST   /api/v1/projects/:id/members   { "user_id": 4, "role": "member" }
GET    /api/v1/projects/:id/members
```

### Tasks

```http
POST   /api/v1/tasks
GET    /api/v1/tasks?project_id=1&status=in_progress&priority=high&sort=due_date&page=1&limit=20
GET    /api/v1/tasks/:id
PUT    /api/v1/tasks/:id
DELETE /api/v1/tasks/:id
```

Допустимые значения `sort`: `created_at`, `created_at_desc`, `due_date`, `priority`.

### Comments

```http
POST /api/v1/tasks/:id/comments   { "content": "..." }
GET  /api/v1/tasks/:id/comments
```

### Stats

```http
GET /api/v1/stats/projects/:id    
GET /api/v1/stats/users/:id
```


## Тесты

```bash
make test         
make test-cover    
```


