# Развертывание Multilayer в Kubernetes

Этот каталог содержит Kubernetes манифесты для развертывания приложения Multilayer в кластер.

## Предварительные требования

1. **Kubernetes кластер** (локальный или удаленный)
2. **kubectl** настроенный для работы с кластером
3. **Docker** для сборки образа
4. **NGINX Ingress Controller** (опционально, для внешнего доступа)

## Структура файлов

```
k8s/
├── namespace.yaml          # Namespace для приложения
├── configmap.yaml          # Конфигурация приложения
├── secret.yaml            # Секретные данные (пароли)
├── postgres-pvc.yaml      # Persistent Volume Claim для БД
├── postgres-deployment.yaml # Deployment PostgreSQL
├── postgres-service.yaml  # Service PostgreSQL
├── app-deployment.yaml    # Deployment основного приложения
├── app-service.yaml       # Service основного приложения
├── ingress.yaml           # Ingress для внешнего доступа
├── deploy.sh              # Скрипт развертывания
├── undeploy.sh            # Скрипт удаления
└── README.md              # Этот файл
```

## Быстрое развертывание

### 1. Сборка Docker образа

```bash
make docker-build-prod
```

### 2. Развертывание в Kubernetes

```bash
make k8s-deploy
```

### 3. Проверка статуса

```bash
make k8s-status
```

### 4. Удаление развертывания

```bash
make k8s-undeploy
```

## Ручное развертывание

Если вы хотите развернуть вручную:

```bash
# 1. Создать namespace
kubectl apply -f namespace.yaml

# 2. Применить конфигурацию
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml

# 3. Развернуть PostgreSQL
kubectl apply -f postgres-pvc.yaml
kubectl apply -f postgres-deployment.yaml
kubectl apply -f postgres-service.yaml

# 4. Развернуть приложение
kubectl apply -f app-deployment.yaml
kubectl apply -f app-service.yaml

# 5. Настроить Ingress (опционально)
kubectl apply -f ingress.yaml
```

## Конфигурация

### Переменные окружения

Основные переменные настраиваются в `configmap.yaml`:

- `ENV`: Окружение (production)
- `DB_TYPE`: Тип базы данных (postgres)
- `DB_HOST`: Хост базы данных
- `DB_PORT`: Порт базы данных
- `DB_NAME`: Имя базы данных
- `PORT`: Порт приложения

### Секретные данные

Чувствительные данные хранятся в `secret.yaml`:

- `DB_USER`: Пользователь базы данных
- `DB_PASSWORD`: Пароль базы данных

**Важно**: В продакшене замените значения в secret.yaml на реальные, закодированные в base64.

## Доступ к приложению

### Через Ingress (рекомендуется)

1. Добавьте в `/etc/hosts`:
   ```
   <CLUSTER_IP> multilayer.local
   ```

2. Откройте в браузере: `http://multilayer.local`

### Через Port Forward

```bash
kubectl port-forward svc/multilayer-app 8080:80 -n multilayer
```

Затем откройте: `http://localhost:8080`

### Через NodePort (если Ingress недоступен)

Измените тип сервиса в `app-service.yaml`:

```yaml
type: NodePort
```

## Мониторинг и логи

### Просмотр логов приложения

```bash
kubectl logs -f deployment/multilayer-app -n multilayer
```

### Просмотр логов PostgreSQL

```bash
kubectl logs -f deployment/multilayer-postgres -n multilayer
```

### Проверка статуса подов

```bash
kubectl get pods -n multilayer
```

### Проверка сервисов

```bash
kubectl get svc -n multilayer
```

## Масштабирование

### Горизонтальное масштабирование приложения

```bash
kubectl scale deployment multilayer-app --replicas=3 -n multilayer
```

### Вертикальное масштабирование (изменение ресурсов)

Отредактируйте `app-deployment.yaml` и примените изменения:

```bash
kubectl apply -f app-deployment.yaml
```

## Устранение неполадок

### Проверка событий

```bash
kubectl get events -n multilayer --sort-by='.lastTimestamp'
```

### Описание подов

```bash
kubectl describe pod <pod-name> -n multilayer
```

### Проверка конфигурации

```bash
kubectl get configmap multilayer-config -n multilayer -o yaml
kubectl get secret multilayer-secret -n multilayer -o yaml
```

## Безопасность

1. **Секреты**: В продакшене используйте Kubernetes Secrets или внешние системы управления секретами
2. **Сеть**: Настройте Network Policies для ограничения трафика
3. **RBAC**: Настройте роли и права доступа
4. **Обновления**: Используйте Rolling Updates для обновления приложения

## Производительность

### Рекомендуемые ресурсы

- **Приложение**: 128Mi RAM, 100m CPU (запросы) / 256Mi RAM, 200m CPU (лимиты)
- **PostgreSQL**: 512Mi RAM, 250m CPU (запросы) / 1Gi RAM, 500m CPU (лимиты)

### Настройка для высоких нагрузок

1. Увеличьте количество реплик приложения
2. Настройте HPA (Horizontal Pod Autoscaler)
3. Используйте кэширование (Redis)
4. Оптимизируйте запросы к базе данных 