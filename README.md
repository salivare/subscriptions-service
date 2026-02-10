# Subscriptions Service

Сервис управления подписками.

## 🚀 Быстрый старт (Docker Compose)

Самый простой способ поднять всё окружение одной командой. Система автоматически запустит базу данных, применит миграции и поднимет сервис.

```bash
docker-compose up --build
```

## Доступ к сервису:
#### API: http://0.0.0.0:8082/
#### Swagger UI: http://localhost:8082/swagger/

## 🛠 Запуск через TaskFile
Для удобства разработки используется Taskfile.
### Установка Task

Если он еще не установлен:

*   **macOS:**
    ```bash
    brew install go-task/tap/go-task
    ```
*   **Linux:**
    ```bash
    sh -c "$(curl --location https://taskfile.dev)" -- -d
    ```
*   **Windows:**
    ```powershell
    choco install go-task
    ```

### Команды

| Команда | Описание |
| :--- | :--- |
| `task run-compose` | Пересборка и запуск проекта через docker-compose |
| `task migrate` | Сборка образа и запуск миграций в тестовую БД |
| `task run-tests` | Запуск интеграционных тестов в контейнере |

## 🧪 Интеграционное тестирование

Чтобы запустить тесты, необходимо, чтобы сервис и БД были запущены. Если тестовая БД отсутствует, она будет создана автоматически.

1. **Примените миграции:**
```bash
task migrate
```

2. Запуск тестов 
```bash
task run-tests
```
## Ручной запуск для тестов
### Миграции
```bash
# Сборка
docker build -f Docker/migrator/Dockerfile -t subscriptions-migrator .

# Запуск
docker run --rm --name test_migrator \
  -e CONFIG_PATH="/config/config.yaml" \
  -v ./configs/test.yaml:/config/config.yaml:ro \
  -v $(pwd)/migrations:/app/migrations:ro \
  subscriptions-migrator

```
### Тесты
```bash
docker run --rm --name test_runner \
  -e CONFIG_PATH="/config/config.yaml" \
  -v $(pwd):/app \
  -v ./configs/test.yaml:/config/config.yaml:ro \
  -w /app \
  golang:1.25.7-alpine3.22 \
  sh -c "apk add --no-cache git && go mod download && go test ./tests/... -v"

```