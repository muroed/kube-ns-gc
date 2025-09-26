# Настройка Telegram уведомлений

## Создание Telegram бота

### 1. Создать бота через BotFather

1. Откройте Telegram и найдите [@BotFather](https://t.me/botfather)
2. Отправьте команду `/newbot`
3. Введите имя для вашего бота (например: "Kubernetes Namespace GC")
4. Введите username для бота (например: "kube_ns_gc_bot")
5. Сохраните полученный токен - он понадобится для настройки

### 2. Получить Chat ID

#### Способ 1: Через бота @userinfobot
1. Найдите бота [@userinfobot](https://t.me/userinfobot)
2. Отправьте ему любое сообщение
3. Бот вернет ваш Chat ID

#### Способ 2: Через API
1. Отправьте сообщение вашему боту
2. Откройте в браузере: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
3. Найдите в ответе `"chat":{"id":123456789}` - это ваш Chat ID

#### Способ 3: Для группового чата
1. Добавьте бота в группу
2. Отправьте сообщение в группу
3. Используйте API как в способе 2
4. Chat ID для группы будет отрицательным числом

## Настройка в kube-ns-gc

### Через Helm values

Создайте файл `telegram-values.yaml`:

```yaml
config:
  telegram:
    enabled: true
    botToken: "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"
    chatId: "123456789"
    parseMode: "Markdown"
```

Установите с настройками:

```bash
helm upgrade --install kube-ns-gc ./helm/kube-ns-gc \
  --namespace kube-ns-gc \
  --values telegram-values.yaml
```

### Через переменные окружения

```bash
helm upgrade --install kube-ns-gc ./helm/kube-ns-gc \
  --namespace kube-ns-gc \
  --set config.telegram.enabled=true \
  --set config.telegram.botToken="1234567890:ABCdefGHIjklMNOpqrsTUVwxyz" \
  --set config.telegram.chatId="123456789" \
  --set config.telegram.parseMode="Markdown"
```

### Через ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-ns-gc-config
  namespace: kube-ns-gc
data:
  config.json: |
    {
      "cleanup_interval": "24h",
      "namespace_max_age": "168h",
      "helm_release_timeout": "5m",
      "excluded_namespaces": ["kube-system", "kube-public", "kube-node-lease", "default"],
      "ignore_label": "kube-ns-gc.ignore",
      "log_level": "info",
      "port": 8080,
      "telegram": {
        "enabled": true,
        "bot_token": "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz",
        "chat_id": "123456789",
        "parse_mode": "Markdown"
      }
    }
```

## Безопасность

### Использование Kubernetes Secrets

Для безопасности рекомендуется хранить токен бота в Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kube-ns-gc-telegram
  namespace: kube-ns-gc
type: Opaque
stringData:
  bot-token: "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"
  chat-id: "123456789"
```

Обновите ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-ns-gc-config
  namespace: kube-ns-gc
data:
  config.json: |
    {
      "cleanup_interval": "24h",
      "namespace_max_age": "168h",
      "helm_release_timeout": "5m",
      "excluded_namespaces": ["kube-system", "kube-public", "kube-node-lease", "default"],
      "ignore_label": "kube-ns-gc.ignore",
      "log_level": "info",
      "port": 8080,
      "telegram": {
        "enabled": true,
        "bot_token": "${TELEGRAM_BOT_TOKEN}",
        "chat_id": "${TELEGRAM_CHAT_ID}",
        "parse_mode": "Markdown"
      }
    }
```

И обновите Deployment для использования переменных окружения:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kube-ns-gc
spec:
  template:
    spec:
      containers:
      - name: kube-ns-gc
        env:
        - name: TELEGRAM_BOT_TOKEN
          valueFrom:
            secretKeyRef:
              name: kube-ns-gc-telegram
              key: bot-token
        - name: TELEGRAM_CHAT_ID
          valueFrom:
            secretKeyRef:
              name: kube-ns-gc-telegram
              key: chat-id
```

## Типы уведомлений

### 1. Уведомление о запуске
```
🚀 kube-ns-gc Started

🕐 Time: 2024-01-15 10:30:00 UTC
📋 Service is now monitoring namespaces for cleanup
```

### 2. Удаление неймспейса
```
🗑️ Namespace Deleted

📦 Namespace: test-namespace
⏰ Age: 2h30m
🕐 Time: 2024-01-15 10:30:00 UTC
```

### 3. Удаление Helm релиза
```
🧹 Helm Release Deleted

📦 Release: my-app
🏠 Namespace: test-namespace
🕐 Time: 2024-01-15 10:30:00 UTC
```

### 4. Сводка очистки
```
📊 Cleanup Summary

🔍 Total namespaces checked: 15
🗑️ Namespaces deleted: 3
⏱️ Cleanup duration: 45s
🕐 Time: 2024-01-15 10:30:00 UTC
```

### 5. Ошибки
```
❌ Error

📝 Message: Failed to delete namespace test-namespace
🔍 Error: namespace "test-namespace" not found
🕐 Time: 2024-01-15 10:30:00 UTC
```

## Тестирование

### Проверка настройки

1. Проверьте логи пода:
```bash
kubectl logs -n kube-ns-gc -l app.kubernetes.io/name=kube-ns-gc
```

2. Должно появиться сообщение о запуске в Telegram

### Ручное тестирование

Создайте тестовый неймспейс:

```bash
kubectl create namespace test-cleanup
```

Если настройки корректны, через указанное время (namespace_max_age) неймспейс будет удален и вы получите уведомление.

## Отключение уведомлений

Чтобы отключить Telegram уведомления:

```bash
helm upgrade kube-ns-gc ./helm/kube-ns-gc \
  --namespace kube-ns-gc \
  --set config.telegram.enabled=false
```

Или установите `enabled: false` в values.yaml.
