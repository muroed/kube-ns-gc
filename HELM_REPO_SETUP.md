# 📦 Настройка Helm репозитория

## Что это?

Helm репозиторий позволяет устанавливать чарты через стандартные команды `helm repo add` и `helm install`.

## Автоматическая настройка

После загрузки проекта на GitHub и создания первого релиза Helm репозиторий настроится автоматически через GitHub Pages.

### Что происходит автоматически:

1. **При пуше в main** или **создании тега** запускается `pages.yml` workflow
2. **Упаковывается Helm чарт** с текущей версией
3. **Создается index.yaml** с метаданными чарта
4. **Развертывается на GitHub Pages** по адресу `https://muroed.github.io/kube-ns-gc`

## Ручная настройка GitHub Pages

Если автоматическая настройка не сработала:

### 1. Включить GitHub Pages

1. Перейдите в **Settings** → **Pages**
2. В разделе **Source** выберите **"GitHub Actions"**
3. Сохраните настройки

### 2. Проверить права доступа

1. Перейдите в **Settings** → **Actions** → **General**
2. В разделе **Workflow permissions** выберите **"Read and write permissions"**
3. Включите **"Allow GitHub Actions to create and approve pull requests"**
4. Нажмите **"Save"**

### 3. Запустить workflow вручную

1. Перейдите в **Actions** → **Deploy Helm Chart to GitHub Pages**
2. Нажмите **"Run workflow"**
3. Выберите ветку **main**
4. Нажмите **"Run workflow"**

## Использование Helm репозитория

### Добавить репозиторий

```bash
helm repo add kube-ns-gc https://muroed.github.io/kube-ns-gc
helm repo update
```

### Проверить доступные версии

```bash
helm search repo kube-ns-gc
```

### Установить чарт

```bash
# Последняя версия
helm install kube-ns-gc kube-ns-gc/kube-ns-gc

# Конкретная версия
helm install kube-ns-gc kube-ns-gc/kube-ns-gc --version 0.1.0
```

### Обновить чарт

```bash
# Проверить доступные версии
helm repo update

# Обновить до последней версии
helm upgrade kube-ns-gc kube-ns-gc/kube-ns-gc

# Обновить до конкретной версии
helm upgrade kube-ns-gc kube-ns-gc/kube-ns-gc --version 0.2.0
```

## Структура репозитория

После настройки GitHub Pages будет доступен по адресу:

```
https://muroed.github.io/kube-ns-gc/
├── index.yaml          # Метаданные всех версий чарта
├── kube-ns-gc-0.1.0.tgz # Чарт версии 0.1.0
├── kube-ns-gc-0.2.0.tgz # Чарт версии 0.2.0
└── ...
```

## Troubleshooting

### Ошибка "repository not found"

```bash
# Проверьте URL репозитория
curl -I https://muroed.github.io/kube-ns-gc/index.yaml

# Должен вернуть 200 OK
```

### Ошибка "chart not found"

```bash
# Обновите кэш репозитория
helm repo update

# Проверьте доступные версии
helm search repo kube-ns-gc --versions
```

### Workflow не запускается

1. Проверьте права доступа в **Settings** → **Actions** → **General**
2. Убедитесь, что GitHub Pages включен в **Settings** → **Pages**
3. Проверьте логи workflow в **Actions**

### GitHub Pages не работает

1. Убедитесь, что в **Settings** → **Pages** выбран **"GitHub Actions"**
2. Проверьте, что workflow завершился успешно
3. Подождите 5-10 минут для распространения изменений

## Мониторинг

### Проверить статус GitHub Pages

- **Settings** → **Pages** → показывает статус развертывания

### Проверить логи workflow

- **Actions** → **Deploy Helm Chart to GitHub Pages** → показывает детали выполнения

### Проверить доступность репозитория

```bash
# Проверить доступность
curl -s https://muroed.github.io/kube-ns-gc/index.yaml | head -20

# Должен показать содержимое index.yaml
```

## Преимущества Helm репозитория

1. **Удобство**: Стандартные команды Helm
2. **Версионирование**: Легко переключаться между версиями
3. **Обновления**: Простое обновление через `helm upgrade`
4. **Кэширование**: Helm кэширует метаданные для быстрого доступа
5. **Автоматизация**: Автоматическое обновление при новых релизах
