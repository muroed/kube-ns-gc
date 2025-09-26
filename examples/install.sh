#!/bin/bash

# Пример установки kube-ns-gc

set -e

NAMESPACE="kube-ns-gc"
RELEASE_NAME="kube-ns-gc"

echo "Installing kube-ns-gc..."

# Создать namespace
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Установить через Helm
helm upgrade --install $RELEASE_NAME ./deploy/kube-ns-gc \
  --namespace $NAMESPACE \
  --set config.namespaceMaxAge=72h \
  --set config.excludedNamespaces[0]=production \
  --set config.excludedNamespaces[1]=kube-system \
  --set config.cleanupInterval=12h \
  --set config.logLevel=info \
  --set config.telegram.enabled=false \
  --set config.telegram.botToken="" \
  --set config.telegram.chatId=""

echo "Installation completed!"
echo "Check the status with: kubectl get pods -n $NAMESPACE"
echo "View logs with: kubectl logs -n $NAMESPACE -l app.kubernetes.io/name=kube-ns-gc"
