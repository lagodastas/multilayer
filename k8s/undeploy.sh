#!/bin/bash

# Скрипт для удаления multilayer из Kubernetes кластера

set -e

echo "🗑️  Начинаем удаление multilayer из Kubernetes..."

# Удаляем Ingress
echo "🌐 Удаляем Ingress..."
kubectl delete -f ingress.yaml --ignore-not-found=true

# Удаляем приложение
echo "🔄 Удаляем основное приложение..."
kubectl delete -f app-service.yaml --ignore-not-found=true
kubectl delete -f app-deployment.yaml --ignore-not-found=true

# Удаляем PostgreSQL
echo "🐘 Удаляем PostgreSQL..."
kubectl delete -f postgres-service.yaml --ignore-not-found=true
kubectl delete -f postgres-deployment.yaml --ignore-not-found=true
kubectl delete -f postgres-pvc.yaml --ignore-not-found=true

# Удаляем конфигурацию
echo "⚙️  Удаляем ConfigMap и Secret..."
kubectl delete -f secret.yaml --ignore-not-found=true
kubectl delete -f configmap.yaml --ignore-not-found=true

# Удаляем namespace
echo "📦 Удаляем namespace..."
kubectl delete -f namespace.yaml --ignore-not-found=true

echo "✅ Удаление завершено!" 