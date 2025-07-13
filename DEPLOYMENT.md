# Развертывание Multilayer

Этот документ содержит инструкции по развертыванию приложения Multilayer в различных средах.

## 📁 Структура Kubernetes манифестов

```
k8s/
├── namespace.yaml          # Namespace для приложения
├── configmap.yaml          # Конфигурация приложения
├── secret.yaml            # Секретные данные
├── postgres-pvc.yaml      # Persistent Volume Claim для БД
├── postgres-deployment.yaml # Deployment PostgreSQL
├── postgres-service.yaml  # Service PostgreSQL
├── app-deployment.yaml    # Deployment основного приложения
├── app-service.yaml       # Service основного приложения
├── ingress.yaml           # Ingress для внешнего доступа
├── deploy.sh              # Скрипт развертывания
├── deploy-local.sh        # Скрипт для локального кластера
├── undeploy.sh            # Скрипт удаления
└── README.md              # Документация
```

## 📋 Команды для развертывания

### Для локального кластера (Docker Desktop, Minikube):
```bash
make k8s-deploy-local
```

### Для продакшен кластера:
```bash
make k8s-deploy
```

### Проверка статуса:
```bash
make k8s-status
```

### Удаление развертывания:
```bash
make k8s-undeploy
```

## 📋 Что включено в конфигурацию

1. **Namespace** `multilayer` для изоляции
2. **ConfigMap** с переменными окружения
3. **Secret** для хранения паролей
4. **PostgreSQL** с Persistent Volume
5. **Приложение** с health checks
6. **Service** для внутреннего доступа
7. **Ingress** для внешнего доступа
8. **Health check** эндпоинт `/health`

## 📋 Предварительные требования

1. **Kubernetes кластер** (локальный или удаленный)
2. **kubectl** настроенный для работы с кластером
3. **Docker** для сборки образа
4. **NGINX Ingress Controller** (опционально)

## 🌐 Доступ к приложению

После развертывания вы сможете получить доступ к приложению:

- **Port Forward**: `kubectl port-forward svc/multilayer-app 8080:80 -n multilayer`
- **Ingress**: `http://multilayer.local` (требует настройки /etc/hosts)

## 📚 Документация

Созданы подробные инструкции в:
- `k8s/README.md` - документация по Kubernetes
- `DEPLOYMENT.md` - общие инструкции по развертыванию

Теперь вы можете развернуть ваш проект в любой Kubernetes кластер! Просто выполните `make k8s-deploy-local` для локального кластера или `make k8s-deploy` для продакшена. 