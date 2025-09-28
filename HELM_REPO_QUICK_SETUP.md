# 🚀 Быстрая настройка Helm репозитория

## После загрузки на GitHub

### 1. Включить GitHub Pages

1. Перейдите в **Settings** → **Pages**
2. В разделе **Source** выберите **"GitHub Actions"**
3. Нажмите **"Save"**

### 2. Проверить права Actions

1. Перейдите в **Settings** → **Actions** → **General**
2. В разделе **Workflow permissions** выберите **"Read and write permissions"**
3. Нажмите **"Save"**

### 3. Создать первый релиз

```bash
# Создать релиз версии 0.1.0
./scripts/release.sh 0.1.0
```

### 4. Проверить Helm репозиторий

```bash
# Добавить репозиторий
helm repo add kube-ns-gc https://muroed.github.io/kube-ns-gc
helm repo update

# Проверить доступные версии
helm search repo kube-ns-gc

# Установить
helm install kube-ns-gc kube-ns-gc/kube-ns-gc --namespace kube-ns-gc --create-namespace
```

## URL репозитория

После настройки Helm репозиторий будет доступен по адресу:

**https://muroed.github.io/kube-ns-gc**

## Что происходит автоматически

- ✅ При пуше в main → обновляется Helm репозиторий
- ✅ При создании тега → создается релиз + обновляется Helm репозиторий
- ✅ Каждая версия чарта соответствует версии Docker образа
- ✅ Security scan запускается автоматически

## Проверка работы

```bash
# Проверить доступность репозитория
curl -I https://muroed.github.io/kube-ns-gc/index.yaml

# Должен вернуть: HTTP/2 200
```

## Troubleshooting

Если что-то не работает:

1. **Проверьте GitHub Pages**: Settings → Pages → должен быть "GitHub Actions"
2. **Проверьте Actions**: должны быть права "Read and write"
3. **Проверьте workflow**: Actions → "Deploy Helm Chart to GitHub Pages"
4. **Подождите 5-10 минут** для распространения изменений

## Готово! 🎉

Теперь пользователи могут устанавливать ваш чарт стандартными командами Helm:

```bash
helm repo add kube-ns-gc https://muroed.github.io/kube-ns-gc
helm install kube-ns-gc kube-ns-gc/kube-ns-gc
```
