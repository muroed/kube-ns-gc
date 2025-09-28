# 🔧 Исправление проблем с GitHub Pages

## Проблемы

### 1. Docker Buildx Cache Error
```
ERROR: failed to build: Cache export is not supported for the docker driver.
Switch to a different driver, or turn on the containerd image store, and try again.
```

### 2. GitHub Pages Environment Protection
```
Tag "v1.1.0" is not allowed to deploy to github-pages due to environment protection rules.
The deployment was rejected or didn't satisfy other protection rules.
```

## Решения

### ✅ 1. Исправление Docker Buildx

**Проблема**: GitHub Actions cache не поддерживается с драйвером `docker`.

**Решение**: Убрали кэширование из build-push-action:

```yaml
# Было (не работает):
- name: Build and push Docker image
  uses: docker/build-push-action@v5
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max

# Стало (работает):
- name: Build and push Docker image
  uses: docker/build-push-action@v5
  with:
    context: .
    file: ./Dockerfile
    platforms: linux/amd64,linux/arm64
    push: true
    tags: |
      ghcr.io/muroed/kube-ns-gc:${{ steps.version.outputs.version }}
      ghcr.io/muroed/kube-ns-gc:latest
```

### ✅ 2. Исправление GitHub Pages

**Проблема**: Environment protection rules блокируют развертывание из тегов.

**Решение**: Убрали environment из pages.yml:

```yaml
# Было (блокируется):
jobs:
  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    runs-on: ubuntu-latest

# Стало (работает):
jobs:
  deploy:
    runs-on: ubuntu-latest
```

## Альтернативные решения

### Для Docker Cache (если нужен кэш)

1. **Переключиться на docker-container драйвер**:
```yaml
- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3
  with:
    driver: docker-container
```

2. **Использовать registry cache**:
```yaml
- name: Build and push Docker image
  uses: docker/build-push-action@v5
  with:
    cache-from: type=registry,ref=ghcr.io/muroed/kube-ns-gc:cache
    cache-to: type=registry,ref=ghcr.io/muroed/kube-ns-gc:cache,mode=max
```

### Для GitHub Pages (если нужны protection rules)

1. **Настроить правила в Settings**:
   - Settings → Environments → github-pages
   - Убрать "Required reviewers" для тегов
   - Добавить правило для тегов `v*`

2. **Использовать отдельный workflow для тегов**:
```yaml
on:
  push:
    branches: [main]
  workflow_dispatch:
```

## Проверка

После исправлений:

1. ✅ **Docker builds** работают без ошибок кэша
2. ✅ **GitHub Pages** развертывается из тегов
3. ✅ **Helm репозиторий** обновляется автоматически

## Мониторинг

- **Actions**: https://github.com/muroed/kube-ns-gc/actions
- **Pages**: https://github.com/muroed/kube-ns-gc/settings/pages
- **Environments**: https://github.com/muroed/kube-ns-gc/settings/environments

## Результат

Теперь релизы работают полностью автоматически:
1. Создание тега → Docker build → GitHub Pages deploy
2. Helm репозиторий обновляется автоматически
3. Нет блокировок по protection rules
