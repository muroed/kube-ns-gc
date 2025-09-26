# Руководство по развертыванию kube-ns-gc

## Предварительные требования

- Kubernetes кластер версии 1.19+
- Helm 3.x
- kubectl настроен для работы с кластером
- Права администратора кластера

## Быстрая установка

### 1. Клонирование репозитория

```bash
git clone https://github.com/your-org/kube-ns-gc.git
cd kube-ns-gc
```

### 2. Установка через Helm

```bash
# Создать namespace
kubectl create namespace kube-ns-gc

# Установить с настройками по умолчанию
helm install kube-ns-gc ./helm/kube-ns-gc \
  --namespace kube-ns-gc \
  --create-namespace
```

### 3. Проверка установки

```bash
# Проверить статус подов
kubectl get pods -n kube-ns-gc

# Проверить логи
kubectl logs -n kube-ns-gc -l app.kubernetes.io/name=kube-ns-gc

# Проверить health check
kubectl port-forward -n kube-ns-gc svc/kube-ns-gc 8080:8080
curl http://localhost:8080/health
```

## Настройка конфигурации

### Через values.yaml

Создайте файл `custom-values.yaml`:

```yaml
config:
  # Запускать очистку каждые 12 часов
  cleanupInterval: "12h"
  
  # Удалять неймспейсы старше 3 дней
  namespaceMaxAge: "72h"
  
  # Таймаут удаления Helm релизов
  helmReleaseTimeout: "10m"
  
  # Исключенные неймспейсы
  excludedNamespaces:
    - kube-system
    - kube-public
    - kube-node-lease
    - default
    - kube-ns-gc
    - production
    - staging
    - monitoring
    - logging
  
  # Лейбл для игнорирования
  ignoreLabel: "kube-ns-gc.ignore"
  
  # Уровень логирования
  logLevel: "info"

# Ресурсы
resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

Установите с кастомными настройками:

```bash
helm upgrade --install kube-ns-gc ./helm/kube-ns-gc \
  --namespace kube-ns-gc \
  --values custom-values.yaml
```

### Через переменные окружения

```bash
helm upgrade --install kube-ns-gc ./helm/kube-ns-gc \
  --namespace kube-ns-gc \
  --set config.namespaceMaxAge=48h \
  --set config.cleanupInterval=6h \
  --set config.excludedNamespaces[0]=production \
  --set config.excludedNamespaces[1]=staging
```

## Защита неймспейсов от удаления

### Способ 1: Добавить в список исключений

```yaml
config:
  excludedNamespaces:
    - important-namespace
    - production
    - staging
```

### Способ 2: Добавить лейбл на неймспейс

```bash
kubectl label namespace my-important-namespace kube-ns-gc.ignore=true
```

### Способ 3: Проверить существующие лейблы

```bash
kubectl get namespaces --show-labels
```

## Мониторинг и логирование

### Просмотр логов

```bash
# Логи в реальном времени
kubectl logs -n kube-ns-gc -l app.kubernetes.io/name=kube-ns-gc -f

# Логи за последний час
kubectl logs -n kube-ns-gc -l app.kubernetes.io/name=kube-ns-gc --since=1h
```

### Метрики

```bash
# Получить метрики
kubectl port-forward -n kube-ns-gc svc/kube-ns-gc 8080:8080
curl http://localhost:8080/metrics
```

### Health Check

```bash
# Проверить здоровье сервиса
kubectl port-forward -n kube-ns-gc svc/kube-ns-gc 8080:8080
curl http://localhost:8080/health
```

## Обновление

### Обновление через Helm

```bash
# Обновить до новой версии
helm upgrade kube-ns-gc ./helm/kube-ns-gc \
  --namespace kube-ns-gc \
  --values custom-values.yaml

# Проверить статус обновления
helm status kube-ns-gc -n kube-ns-gc
```

### Откат изменений

```bash
# Посмотреть историю релизов
helm history kube-ns-gc -n kube-ns-gc

# Откатиться к предыдущей версии
helm rollback kube-ns-gc 1 -n kube-ns-gc
```

## Удаление

### Полное удаление

```bash
# Удалить Helm релиз
helm uninstall kube-ns-gc -n kube-ns-gc

# Удалить namespace (опционально)
kubectl delete namespace kube-ns-gc
```

## Устранение неполадок

### Проблема: Под не запускается

```bash
# Проверить события
kubectl describe pod -n kube-ns-gc -l app.kubernetes.io/name=kube-ns-gc

# Проверить логи
kubectl logs -n kube-ns-gc -l app.kubernetes.io/name=kube-ns-gc
```

### Проблема: Нет прав на удаление неймспейсов

```bash
# Проверить RBAC
kubectl auth can-i delete namespaces --as=system:serviceaccount:kube-ns-gc:kube-ns-gc

# Проверить ClusterRole
kubectl describe clusterrole kube-ns-gc
```

### Проблема: Helm релизы не удаляются

```bash
# Проверить права на secrets (Helm хранит релизы в secrets)
kubectl auth can-i get secrets --as=system:serviceaccount:kube-ns-gc:kube-ns-gc

# Проверить наличие Helm релизов
helm list --all-namespaces
```

## Безопасность

### Рекомендации по безопасности

1. **Ограничьте права**: Микросервис имеет минимальные необходимые права
2. **Мониторинг**: Настройте алерты на удаление неймспейсов
3. **Бэкапы**: Убедитесь, что важные данные имеют бэкапы
4. **Тестирование**: Протестируйте на dev/staging окружениях

### Настройка мониторинга

```yaml
# Prometheus ServiceMonitor
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kube-ns-gc
  namespace: kube-ns-gc
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: kube-ns-gc
  endpoints:
  - port: http
    path: /metrics
```

## Примеры использования

### Сценарий 1: Очистка dev окружения

```yaml
config:
  namespaceMaxAge: "24h"  # Удалять dev неймспейсы через 1 день
  cleanupInterval: "6h"   # Проверять каждые 6 часов
  excludedNamespaces:
    - production
    - staging
    - monitoring
```

### Сценарий 2: Очистка CI/CD неймспейсов

```yaml
config:
  namespaceMaxAge: "2h"   # Удалять CI неймспейсы через 2 часа
  cleanupInterval: "30m"  # Проверять каждые 30 минут
  excludedNamespaces:
    - production
    - staging
    - monitoring
    - logging
```

### Сценарий 3: Консервативная очистка

```yaml
config:
  namespaceMaxAge: "720h" # Удалять через 30 дней
  cleanupInterval: "24h"  # Проверять раз в день
  excludedNamespaces:
    - kube-system
    - kube-public
    - kube-node-lease
    - default
    - kube-ns-gc
    - production
    - staging
    - monitoring
    - logging
    - ingress-nginx
    - cert-manager
```
