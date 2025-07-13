#!/bin/bash

# Скрипт для развертывания multilayer в локальный Kubernetes кластер

set -e

echo "🚀 Начинаем развертывание multilayer в локальный Kubernetes кластер..."

# Проверяем, что kubectl доступен
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl не найден. Установите kubectl и настройте кластер."
    exit 1
fi

# Проверяем подключение к кластеру
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Не удается подключиться к Kubernetes кластеру."
    echo "Убедитесь, что кластер запущен (minikube start, docker desktop, kind и т.д.)"
    exit 1
fi

echo "✅ Подключение к кластеру установлено"

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

# Определяем способ доступа к приложению
if kubectl get ingressclass nginx &> /dev/null; then
    echo "🌐 Настраиваем Ingress..."
    kubectl apply -f ingress.yaml
    echo ""
    echo "🔗 Ingress:"
    kubectl get ingress -n multilayer
    echo ""
    echo "💡 Для доступа к приложению добавьте в /etc/hosts:"
    echo "   $(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}') multilayer.local"
    echo ""
    echo "🌐 Затем откройте: http://multilayer.local"
else
    echo "🌐 Ingress Controller не найден, используем Port Forward"
    echo ""
    echo "🔗 Для доступа к приложению выполните:"
    echo "   kubectl port-forward svc/multilayer-app 8080:80 -n multilayer"
    echo ""
    echo "🌐 Затем откройте: http://localhost:8080"
fi

echo ""
echo "🔍 Для просмотра логов:"
echo "   kubectl logs -f deployment/multilayer-app -n multilayer"
echo ""
echo "🧪 Тестирование API:"
echo "   curl -X POST http://localhost:8080/users -H 'Content-Type: application/json' -d '{\"username\":\"test\",\"email\":\"test@example.com\"}'" 