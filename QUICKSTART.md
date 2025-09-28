# Быстрый старт kube-ns-gc

## Что это?

Микросервис для автоматического удаления старых неймспейсов и Helm релизов в Kubernetes.

## Установка за 5 минут

### 1. Клонировать репозиторий
```bash
git clone https://github.com/your-org/kube-ns-gc.git
cd kube-ns-gc
```

### 2. Установить

#### Вариант A: Через Helm репозиторий (рекомендуется)
```bash
# Добавить репозиторий
helm repo add kube-ns-gc https://muroed.github.io/kube-ns-gc
helm repo update

# Установить
helm install kube-ns-gc kube-ns-gc/kube-ns-gc \
  --namespace kube-ns-gc \
  --create-namespace
```

#### Вариант B: Через GitHub Releases
```bash
# Скачать последний релиз
helm install kube-ns-gc https://github.com/muroed/kube-ns-gc/releases/latest/download/kube-ns-gc-0.1.0.tgz \
  --namespace kube-ns-gc \
  --create-namespace
```

#### Вариант C: Локальная установка
```bash
# Простая установка
./examples/install.sh

# Или через Helm напрямую
helm install kube-ns-gc ./deploy/kube-ns-gc \
  --namespace kube-ns-gc \
  --create-namespace
```

### 3. Проверить
```bash
kubectl get pods -n kube-ns-gc
kubectl logs -n kube-ns-gc -l app.kubernetes.io/name=kube-ns-gc
```

## Основные настройки

### Удалять неймспейсы старше 3 дней
```bash
helm upgrade kube-ns-gc ./deploy/kube-ns-gc \
  --namespace kube-ns-gc \
  --set config.namespaceMaxAge=72h
```

### Проверять каждые 6 часов
```bash
helm upgrade kube-ns-gc ./deploy/kube-ns-gc \
  --namespace kube-ns-gc \
  --set config.cleanupInterval=6h
```

### Исключить неймспейсы
```bash
helm upgrade kube-ns-gc ./deploy/kube-ns-gc \
  --namespace kube-ns-gc \
  --set config.excludedNamespaces[0]=production \
  --set config.excludedNamespaces[1]=staging
```

### Включить Telegram уведомления
```bash
helm upgrade kube-ns-gc ./deploy/kube-ns-gc \
  --namespace kube-ns-gc \
  --set config.telegram.enabled=true \
  --set config.telegram.botToken="YOUR_BOT_TOKEN" \
  --set config.telegram.chatId="YOUR_CHAT_ID" \
  --set config.telegram.notifications.startup=true \
  --set config.telegram.notifications.namespaceDeleted=true
```

## Защита неймспейсов

### Способ 1: Лейбл
```bash
kubectl label namespace my-namespace kube-ns-gc.ignore=true
```

### Способ 2: Список исключений
```yaml
config:
  excludedNamespaces:
    - production
    - staging
    - important-namespace
```

## Мониторинг

```bash
# Health check
kubectl port-forward -n kube-ns-gc svc/kube-ns-gc 8080:8080
curl http://localhost:8080/health

# Метрики
curl http://localhost:8080/metrics
```

## Удаление

```bash
helm uninstall kube-ns-gc -n kube-ns-gc
kubectl delete namespace kube-ns-gc
```

## Поддержка

- 📖 [Полная документация](README.md)
- 🚀 [Руководство по развертыванию](examples/deployment-guide.md)
- 🐛 [Issues](https://github.com/your-org/kube-ns-gc/issues)
