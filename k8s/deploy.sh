#!/bin/bash

# Скрипт для развертывания multilayer в Kubernetes кластер

set -e

echo "🚀 Начинаем развертывание multilayer в Kubernetes..."

# Создаем namespace
echo "📦 Создаем namespace..."
kubectl apply -f namespace.yaml

# Применяем конфигурацию
echo "⚙️  Применяем ConfigMap и Secret..."
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml

# Развертываем PostgreSQL
echo "🐘 Развертываем PostgreSQL..."
kubectl apply -f postgres-pvc.yaml
kubectl apply -f postgres-deployment.yaml
kubectl apply -f postgres-service.yaml

# Ждем готовности PostgreSQL
echo "⏳ Ждем готовности PostgreSQL..."
kubectl wait --for=condition=ready pod -l app=multilayer-postgres -n multilayer --timeout=300s

# Развертываем приложение
echo "🔄 Развертываем основное приложение..."
kubectl apply -f app-deployment.yaml
kubectl apply -f app-service.yaml

# Применяем Ingress
echo "🌐 Настраиваем Ingress..."
kubectl apply -f ingress.yaml

# Ждем готовности приложения
echo "⏳ Ждем готовности приложения..."
kubectl wait --for=condition=ready pod -l app=multilayer-app -n multilayer --timeout=300s

echo "✅ Развертывание завершено!"
echo ""
echo "📊 Статус развертывания:"
kubectl get pods -n multilayer
echo ""
echo "🌐 Сервисы:"
kubectl get svc -n multilayer
echo ""
echo "🔗 Ingress:"
kubectl get ingress -n multilayer
echo ""
echo "💡 Для доступа к приложению добавьте в /etc/hosts:"
echo "   <CLUSTER_IP> multilayer.local"
echo ""
echo "🔍 Для просмотра логов:"
echo "   kubectl logs -f deployment/multilayer-app -n multilayer" 