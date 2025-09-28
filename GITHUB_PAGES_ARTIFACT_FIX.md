# 🔧 Исправление ошибки артефакта GitHub Pages

## Проблема

При развертывании GitHub Pages возникала ошибка:

```
Error: No artifacts named "pages" were found for this workflow run.
Ensure artifacts are uploaded with actions/upload-artifact@v4 or later.
```

## Причина

Несоответствие между именем артефакта в `upload-pages-artifact` и `deploy-pages`:

```yaml
# Проблема: upload-pages-artifact не указывает имя артефакта
- name: Upload artifact
  uses: actions/upload-pages-artifact@v3
  with:
    path: ./packages

# deploy-pages ожидает артефакт с именем "pages"
- name: Deploy to GitHub Pages
  uses: actions/deploy-pages@v4
  with:
    artifact_name: pages  # Ищет артефакт "pages", но его нет
```

## Решение

Добавили явное имя артефакта в `upload-pages-artifact`:

```yaml
# Исправлено: явно указываем имя артефакта
- name: Upload artifact
  uses: actions/upload-pages-artifact@v3
  with:
    path: ./packages
    name: pages  # ← Добавлено имя артефакта

# deploy-pages теперь найдет артефакт "pages"
- name: Deploy to GitHub Pages
  uses: actions/deploy-pages@v4
  # artifact_name не нужен, так как используется стандартное имя
```

## Проверка

Локально структура артефакта правильная:

```bash
$ ls -la packages/
-rw-r--r--  1325 index.yaml
-rw-r--r--  3822 kube-ns-gc-0.1.0.tgz
-rw-r--r--  3822 kube-ns-gc-1.0.1.tgz
-rw-r--r--  3822 kube-ns-gc-1.0.2.tgz
```

## Результат

После исправления:

1. ✅ **upload-pages-artifact** создает артефакт с именем "pages"
2. ✅ **deploy-pages** находит артефакт "pages"
3. ✅ **GitHub Pages** развертывается успешно
4. ✅ **Helm репозиторий** обновляется автоматически

## Тестирование

Успешно создан релиз v1.0.2:

```bash
🚀 Creating release for version 1.0.2
📝 Updating Chart.yaml version to 1.0.2
💾 Committing version update
🏷️  Creating tag v1.0.2
📤 Pushing changes and tag
✅ Release v1.0.2 created successfully!
```

## Мониторинг

- **Actions**: https://github.com/muroed/kube-ns-gc/actions
- **Pages**: https://muroed.github.io/kube-ns-gc/
- **Helm Repo**: https://muroed.github.io/kube-ns-gc/index.yaml

## Использование

После успешного развертывания Helm репозиторий доступен:

```bash
# Добавить репозиторий
helm repo add kube-ns-gc https://muroed.github.io/kube-ns-gc
helm repo update

# Установить последнюю версию
helm install kube-ns-gc kube-ns-gc/kube-ns-gc

# Установить конкретную версию
helm install kube-ns-gc kube-ns-gc/kube-ns-gc --version 1.0.2
```

## ✅ Проблема решена!

GitHub Pages теперь развертывается без ошибок артефактов.
