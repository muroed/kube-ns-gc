# Настройка типов Telegram уведомлений

## Обзор

Теперь вы можете гибко настраивать, какие типы уведомлений отправлять в Telegram. Каждый тип уведомлений можно включить или отключить независимо.

## Доступные типы уведомлений

### 1. 🚀 Startup (Запуск сервиса)
- **Параметр**: `telegram.notifications.startup`
- **По умолчанию**: `true`
- **Описание**: Уведомление о запуске микросервиса

### 2. 🗑️ Namespace Deleted (Удаление неймспейса)
- **Параметр**: `telegram.notifications.namespaceDeleted`
- **По умолчанию**: `true`
- **Описание**: Уведомление о каждом удаленном неймспейсе

### 3. 🧹 Helm Release Deleted (Удаление Helm релиза)
- **Параметр**: `telegram.notifications.helmReleaseDeleted`
- **По умолчанию**: `true`
- **Описание**: Уведомление о каждом удаленном Helm релизе

### 4. 📊 Cleanup Summary (Сводка очистки)
- **Параметр**: `telegram.notifications.cleanupSummary`
- **По умолчанию**: `true`
- **Описание**: Сводка после каждого цикла очистки

### 5. ❌ Errors (Ошибки)
- **Параметр**: `telegram.notifications.errors`
- **По умолчанию**: `true`
- **Описание**: Уведомления об ошибках

## Примеры конфигурации

### Минимальная конфигурация (только ошибки)
```yaml
config:
  telegram:
    enabled: true
    botToken: "YOUR_BOT_TOKEN"
    chatId: "YOUR_CHAT_ID"
    notifications:
      startup: false
      namespaceDeleted: false
      helmReleaseDeleted: false
      cleanupSummary: false
      errors: true  # Только ошибки
```

### Только важные события
```yaml
config:
  telegram:
    enabled: true
    botToken: "YOUR_BOT_TOKEN"
    chatId: "YOUR_CHAT_ID"
    notifications:
      startup: true
      namespaceDeleted: true
      helmReleaseDeleted: false
      cleanupSummary: true
      errors: true
```

### Только сводки (без детализации)
```yaml
config:
  telegram:
    enabled: true
    botToken: "YOUR_BOT_TOKEN"
    chatId: "YOUR_CHAT_ID"
    notifications:
      startup: true
      namespaceDeleted: false
      helmReleaseDeleted: false
      cleanupSummary: true
      errors: true
```

### Полная детализация
```yaml
config:
  telegram:
    enabled: true
    botToken: "YOUR_BOT_TOKEN"
    chatId: "YOUR_CHAT_ID"
    notifications:
      startup: true
      namespaceDeleted: true
      helmReleaseDeleted: true
      cleanupSummary: true
      errors: true
```

## Настройка через Helm

### Через values.yaml
```bash
helm upgrade --install kube-ns-gc ./deploy/kube-ns-gc \
  --namespace kube-ns-gc \
  --values custom-values.yaml
```

### Через --set параметры
```bash
helm upgrade --install kube-ns-gc ./deploy/kube-ns-gc \
  --namespace kube-ns-gc \
  --set config.telegram.enabled=true \
  --set config.telegram.botToken="YOUR_BOT_TOKEN" \
  --set config.telegram.chatId="YOUR_CHAT_ID" \
  --set config.telegram.notifications.startup=true \
  --set config.telegram.notifications.namespaceDeleted=false \
  --set config.telegram.notifications.helmReleaseDeleted=false \
  --set config.telegram.notifications.cleanupSummary=true \
  --set config.telegram.notifications.errors=true
```

## Настройка через переменные окружения

```bash
export TELEGRAM_ENABLED=true
export TELEGRAM_BOT_TOKEN="YOUR_BOT_TOKEN"
export TELEGRAM_CHAT_ID="YOUR_CHAT_ID"
export TELEGRAM_NOTIFY_STARTUP=true
export TELEGRAM_NOTIFY_NAMESPACE_DELETED=false
export TELEGRAM_NOTIFY_HELM_RELEASE_DELETED=false
export TELEGRAM_NOTIFY_CLEANUP_SUMMARY=true
export TELEGRAM_NOTIFY_ERRORS=true
```

## Рекомендации по настройке

### Для production окружения
```yaml
notifications:
  startup: true          # Важно знать о перезапусках
  namespaceDeleted: false # Может быть слишком много уведомлений
  helmReleaseDeleted: false # Может быть слишком много уведомлений
  cleanupSummary: true   # Полезная сводка
  errors: true          # Критически важно
```

### Для development окружения
```yaml
notifications:
  startup: true          # Полезно для отладки
  namespaceDeleted: true # Полезно для отслеживания
  helmReleaseDeleted: true # Полезно для отслеживания
  cleanupSummary: true   # Полезная сводка
  errors: true          # Важно для отладки
```

### Для CI/CD окружения
```yaml
notifications:
  startup: false         # Частые перезапуски
  namespaceDeleted: false # Много временных неймспейсов
  helmReleaseDeleted: false # Много временных релизов
  cleanupSummary: true   # Полезная сводка
  errors: true          # Критически важно
```

## Динамическое изменение настроек

Настройки можно изменить без перезапуска пода, обновив ConfigMap:

```bash
kubectl patch configmap kube-ns-gc-config -n kube-ns-gc --type merge -p '{
  "data": {
    "config.json": "{\"telegram\":{\"notifications\":{\"startup\":false,\"namespaceDeleted\":true}}}"
  }
}'
```

Затем перезапустите под для применения изменений:

```bash
kubectl rollout restart deployment kube-ns-gc -n kube-ns-gc
```

## Мониторинг настроек

Проверить текущие настройки можно через логи:

```bash
kubectl logs -n kube-ns-gc -l app.kubernetes.io/name=kube-ns-gc | grep "Telegram"
```

Или через метрики (если добавить соответствующие метрики в будущем).
