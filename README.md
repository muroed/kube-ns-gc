# Kubernetes Namespace Garbage Collector

Микросервис для автоматического удаления старых неймспейсов и Helm релизов в Kubernetes кластере.

## Возможности

- 🗑️ Автоматическое удаление неймспейсов старше N дней
- 🧹 Очистка Helm релизов перед удалением неймспейсов
- ⚙️ Гибкая конфигурация через ConfigMap
- 🏷️ Поддержка исключений и лейблов для игнорирования
- 📊 Метрики и health checks
- 📱 Telegram уведомления о удаляемых неймспейсах и Helm релизах
- 🔒 Безопасность: запуск от непривилегированного пользователя

## Конфигурация

Микросервис настраивается через ConfigMap со следующими параметрами:

| Параметр | Описание | По умолчанию |
|----------|----------|--------------|
| `cleanup_interval` | Периодичность запуска очистки | `24h` |
| `namespace_max_age` | Максимальный возраст неймспейса | `168h` (7 дней) |
| `helm_release_timeout` | Таймаут удаления Helm релиза | `5m` |
| `excluded_namespaces` | Список исключенных неймспейсов | `kube-system`, `kube-public`, `kube-node-lease`, `default` |
| `ignore_label` | Лейбл для игнорирования неймспейса | `kube-ns-gc.ignore` |
| `log_level` | Уровень логирования | `info` |
| `port` | Порт HTTP сервера | `8080` |
| `telegram.enabled` | Включить Telegram уведомления | `false` |
| `telegram.bot_token` | Токен Telegram бота | `""` |
| `telegram.chat_id` | ID чата для уведомлений | `""` |
| `telegram.parse_mode` | Режим форматирования сообщений | `Markdown` |
| `telegram.notifications.startup` | Уведомления о запуске | `true` |
| `telegram.notifications.namespace_deleted` | Уведомления об удалении неймспейсов | `true` |
| `telegram.notifications.helm_release_deleted` | Уведомления об удалении Helm релизов | `true` |
| `telegram.notifications.cleanup_summary` | Сводка очистки | `true` |
| `telegram.notifications.errors` | Уведомления об ошибках | `true` |

## Установка

### Через Helm репозиторий (рекомендуется)

```bash
# Добавить репозиторий
helm repo add kube-ns-gc https://muroed.github.io/kube-ns-gc
helm repo update

# Установить последнюю версию
helm install kube-ns-gc kube-ns-gc/kube-ns-gc \
  --namespace kube-ns-gc \
  --create-namespace \
  --set config.namespaceMaxAge=72h \
  --set config.excludedNamespaces[0]=production

# Установить конкретную версию
helm install kube-ns-gc kube-ns-gc/kube-ns-gc \
  --version 0.1.0 \
  --namespace kube-ns-gc \
  --create-namespace
```

### Через GitHub Releases

```bash
# Скачать чарт из релиза
helm install kube-ns-gc https://github.com/muroed/kube-ns-gc/releases/download/v0.1.0/kube-ns-gc-0.1.0.tgz \
  --namespace kube-ns-gc --create-namespace
```

### Через kubectl

```bash
# Применить манифесты
kubectl apply -f https://raw.githubusercontent.com/your-org/kube-ns-gc/main/deploy/
```

## Использование

### Исключение неймспейса от удаления

1. **Через список исключений** (в конфигурации):
```yaml
config:
  excludedNamespaces:
    - production
    - staging
    - important-namespace
```

2. **Через лейбл** (на неймспейсе):
```bash
kubectl label namespace my-namespace kube-ns-gc.ignore=true
```

### Мониторинг

Микросервис предоставляет следующие эндпоинты:

- `GET /health` - Health check
- `GET /metrics` - Метрики работы

Пример метрик:
```json
{
  "total_namespaces": 15,
  "old_namespaces": 3,
  "excluded_namespaces": 4,
  "cleanup_interval": "24h",
  "namespace_max_age": "168h"
}
```

## Telegram уведомления

Микросервис поддерживает отправку уведомлений в Telegram о:
- 🚀 Запуске сервиса
- 🗑️ Удалении неймспейсов
- 🧹 Удалении Helm релизов
- 📊 Сводке очистки
- ❌ Ошибках

### Быстрая настройка

1. Создайте бота через [@BotFather](https://t.me/botfather)
2. Получите Chat ID через [@userinfobot](https://t.me/userinfobot)
3. Настройте уведомления:

```bash
helm upgrade kube-ns-gc ./deploy/kube-ns-gc \
  --namespace kube-ns-gc \
  --set config.telegram.enabled=true \
  --set config.telegram.botToken="YOUR_BOT_TOKEN" \
  --set config.telegram.chatId="YOUR_CHAT_ID" \
  --set config.telegram.notifications.startup=true \
  --set config.telegram.notifications.namespaceDeleted=true \
  --set config.telegram.notifications.helmReleaseDeleted=false
```

📖 [Подробная инструкция по настройке Telegram](examples/telegram-setup.md)

## Разработка

### Требования

- Go 1.21+
- Docker
- Helm 3.x
- kubectl

### Локальная разработка

```bash
# Клонировать репозиторий
git clone https://github.com/your-org/kube-ns-gc.git
cd kube-ns-gc

# Установить зависимости
go mod download

# Запустить тесты
go test ./...

# Собрать образ
docker build -t kube-ns-gc:latest .

# Запустить локально (требует kubeconfig)
go run .
```

### Переменные окружения

Для локальной разработки можно использовать переменные окружения:

```bash
export CLEANUP_INTERVAL=1h
export NAMESPACE_MAX_AGE=24h
export EXCLUDED_NAMESPACES=default,kube-system
export IGNORE_LABEL=kube-ns-gc.ignore
export LOG_LEVEL=debug
export PORT=8080
```

## Версионирование

Проект использует семантическое версионирование (Semantic Versioning):

- **Версия Helm чарта** = **Версия Docker образа**
- Каждой версии чарта соответствует свой тег образа
- Версия указывается в `deploy/kube-ns-gc/Chart.yaml`

### Создание релиза

```bash
# Используйте скрипт для создания релиза
./scripts/release.sh 1.0.0

# Или вручную
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Автоматизация

При создании git тега `v*`:
1. Обновляется версия в Chart.yaml
2. Собирается Docker образ с тегом версии
3. Упаковывается Helm чарт
4. Создается GitHub Release
5. Запускается security scan

## Безопасность

- Микросервис запускается от непривилегированного пользователя (UID 1001)
- Использует минимальные RBAC права
- Read-only root filesystem
- Security context с ограниченными capabilities
- Автоматическое сканирование уязвимостей с помощью Trivy

## Лицензия

MIT License

## Вклад в проект

1. Fork репозитория
2. Создать feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Создать Pull Request
