# 📋 Система версионирования

## Принципы

Проект использует **синхронизированное версионирование** между Helm чартом и Docker образом:

- ✅ **Версия Helm чарта** = **Тег Docker образа**
- ✅ Каждой версии чарта соответствует свой тег образа
- ✅ Версия указывается в `deploy/kube-ns-gc/Chart.yaml`

## Структура версий

```
Chart.yaml:
├── version: 0.1.0      # Версия чарта
└── appVersion: "0.1.0" # Версия приложения

Docker образ:
└── ghcr.io/muroed/kube-ns-gc:0.1.0  # Тег = версия чарта
```

## Создание релиза

### Автоматический способ (рекомендуется)

```bash
# Используйте скрипт для создания релиза
./scripts/release.sh 1.0.0
```

Скрипт автоматически:
1. Проверяет формат версии (X.Y.Z)
2. Обновляет Chart.yaml
3. Создает git тег
4. Пушит изменения

### Ручной способ

```bash
# 1. Обновите версию в Chart.yaml
# 2. Создайте тег
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

## Автоматизация

При создании git тега `v*` запускается workflow `release.yml`:

### 1. Обновление версии
```yaml
- Обновляет version в Chart.yaml
- Обновляет appVersion в Chart.yaml
```

### 2. Сборка образа
```yaml
- Собирает Docker образ с тегом версии
- Создает дополнительный тег latest
```

### 3. Упаковка чарта
```yaml
- Упаковывает Helm чарт
- Создает индекс репозитория
```

### 4. Создание релиза
```yaml
- Создает GitHub Release
- Прикрепляет упакованный чарт
```

### 5. Security scan
```yaml
- Сканирует образ на уязвимости
- Загружает результаты в Security tab
```

## Теги образов

Для каждой версии создаются следующие теги:

```bash
ghcr.io/muroed/kube-ns-gc:0.1.0           # Основной тег версии
ghcr.io/muroed/kube-ns-gc:0.1.0-abc1234   # Тег с git sha
ghcr.io/muroed/kube-ns-gc:latest          # Latest тег (только для main)
```

## Использование

### Установка из Helm репозитория (рекомендуется)

```bash
# Добавить репозиторий
helm repo add kube-ns-gc https://muroed.github.io/kube-ns-gc
helm repo update

# Установить последнюю версию
helm install kube-ns-gc kube-ns-gc/kube-ns-gc

# Установить конкретную версию
helm install kube-ns-gc kube-ns-gc/kube-ns-gc --version 0.1.0
```

### Установка из GitHub Releases

```bash
helm install kube-ns-gc https://github.com/muroed/kube-ns-gc/releases/download/v0.1.0/kube-ns-gc-0.1.0.tgz
```

### Обновление

```bash
# Обновить до последней версии
helm upgrade kube-ns-gc kube-ns-gc/kube-ns-gc

# Обновить до конкретной версии
helm upgrade kube-ns-gc kube-ns-gc/kube-ns-gc --version 0.2.0
```

## Преимущества

1. **Согласованность**: Версия чарта всегда соответствует версии образа
2. **Трассируемость**: Легко отследить, какой образ используется в каком релизе
3. **Безопасность**: Каждая версия проходит security scan
4. **Автоматизация**: Минимум ручной работы при создании релизов
5. **Семантическое версионирование**: Понятная схема версий

## Примеры версий

```bash
# Patch версия (исправления)
./scripts/release.sh 1.0.1

# Minor версия (новые функции)
./scripts/release.sh 1.1.0

# Major версия (breaking changes)
./scripts/release.sh 2.0.0
```

## Мониторинг

После создания релиза можно отслеживать прогресс:

1. **GitHub Actions**: https://github.com/muroed/kube-ns-gc/actions
2. **Releases**: https://github.com/muroed/kube-ns-gc/releases
3. **Packages**: https://github.com/muroed/kube-ns-gc/pkgs/container/kube-ns-gc
4. **Security**: https://github.com/muroed/kube-ns-gc/security
