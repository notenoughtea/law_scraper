# ✅ Реализовано: Деплой на 77.105.133.231 и CI/CD через GitHub

## 🎯 Что было сделано

Настроен полный CI/CD pipeline для автоматического деплоя на сервер `77.105.133.231` через GitHub Actions.

---

## 📦 Созданные компоненты

### 1. GitHub Actions Workflow

**Файл:** `.github/workflows/deploy.yml`

**Функции:**
- ✅ Автоматический запуск при push в `main`/`master`
- ✅ Запуск тестов перед деплоем
- ✅ Сборка приложения
- ✅ Копирование файлов на сервер по SSH
- ✅ Запуск Docker Compose на сервере
- ✅ Проверка успешности деплоя
- ✅ Ручной запуск через GitHub UI

### 2. Скрипты деплоя

| Скрипт | Назначение |
|--------|-----------|
| `deployment/scripts/setup-server.sh` | Первоначальная настройка сервера (Docker, Docker Compose) |
| `deployment/scripts/deploy.sh` | Деплой приложения на сервер |
| `deployment/scripts/status.sh` | Проверка статуса приложения |
| `deployment/scripts/logs.sh` | Просмотр логов с сервера |
| `deployment/scripts/restart.sh` | Перезапуск приложения |

Все скрипты автоматически делают chmod +x при запуске через Makefile.

### 3. Makefile команды

```bash
# Деплой
make setup-server      # Первоначальная настройка сервера
make deploy           # Деплой на продакшн
make status           # Проверка статуса
make logs-server      # Просмотр логов
make logs-server-follow  # Логи в реальном времени
make restart-server   # Перезапуск приложения
make info            # Информация о проекте
```

### 4. Docker конфигурация

**Файл:** `deployment/docker-compose.prod.yml`

**Особенности:**
- Автоматический перезапуск контейнеров
- Health checks
- Ограничения ресурсов (CPU: 1 core, Memory: 512M)
- Ротация логов (10MB, 5 файлов)

### 5. Документация

| Файл | Описание |
|------|----------|
| `DEPLOYMENT.md` | Полная инструкция по деплою и CI/CD (450+ строк) |
| `DEPLOYMENT_QUICKSTART.md` | Быстрый старт за 5 минут |
| `README.md` | Обновлен с информацией о деплое |
| `CHANGELOG.md` | Версия 2.2.0 с деплоем |

---

## 🚀 Как использовать

### Первоначальная настройка (один раз)

#### 1. Настройка SSH

```bash
# Создать SSH ключ
ssh-keygen -t rsa -b 4096

# Скопировать на сервер
ssh-copy-id root@77.105.133.231

# Проверить подключение
ssh root@77.105.133.231
```

#### 2. Настройка сервера

```bash
# Автоматическая установка Docker и Docker Compose
make setup-server
```

#### 3. Настройка GitHub Secrets

Перейти: `GitHub Repo → Settings → Secrets and variables → Actions`

Добавить секреты:

| Secret | Значение |
|--------|----------|
| `SSH_PRIVATE_KEY` | Содержимое `~/.ssh/id_rsa` |
| `SERVER_HOST` | `77.105.133.231` |
| `SERVER_USER` | `root` |
| `API_URL` | `https://regulation.gov.ru/api/...` |
| `KEYWORDS` | `транспорт,концессии,закупки` |
| `TELEGRAM_BOT_TOKEN` | Токен от BotFather |
| `TELEGRAM_CHAT_ID` | ID чата |
| `CRON_SCHEDULE` | `0 9 * * *` |

### Автоматический деплой через GitHub

```bash
# 1. Внесите изменения в код
git add .
git commit -m "Ваши изменения"

# 2. Push в GitHub
git push origin main

# 3. GitHub Actions автоматически:
#    - Запустит тесты
#    - Соберет приложение
#    - Задеплоит на сервер 77.105.133.231
#    - Перезапустит контейнеры
```

**Проверка:**
- GitHub → Actions → "Deploy to Production"

### Ручной деплой

```bash
# Убедитесь что .env настроен
make check-env

# Деплой
make deploy

# Проверка
make status

# Логи
make logs-server
```

---

## 📊 Мониторинг

### Проверка статуса

```bash
# Через Makefile
make status

# Вывод:
# 🐳 Статус Docker контейнеров
# 💾 Использование диска
# 📁 Размер директории
# 📋 Последние логи
```

### Просмотр логов

```bash
# Последние 100 строк
make logs-server

# В реальном времени
make logs-server-follow

# На сервере
ssh root@77.105.133.231 'cd /opt/law_scraper && docker-compose logs -f'
```

### Управление

```bash
# Перезапуск
make restart-server

# Остановка (на сервере)
ssh root@77.105.133.231 'cd /opt/law_scraper && docker-compose down'

# Запуск (на сервере)
ssh root@77.105.133.231 'cd /opt/law_scraper && docker-compose up -d'
```

---

## 🏗️ Архитектура деплоя

```
┌──────────────────────────────────────────────────────┐
│              Локальная разработка                    │
│                                                      │
│  git push origin main                               │
└───────────────────┬──────────────────────────────────┘
                    │
                    ▼
┌──────────────────────────────────────────────────────┐
│              GitHub Actions                          │
│                                                      │
│  1. Run tests                                       │
│  2. Build application                               │
│  3. SSH to server                                   │
│  4. Copy files                                      │
│  5. Docker Compose up                               │
│  6. Verify deployment                               │
└───────────────────┬──────────────────────────────────┘
                    │
                    ▼
┌──────────────────────────────────────────────────────┐
│         Сервер 77.105.133.231                       │
│         /opt/law_scraper                            │
│                                                      │
│  ┌────────────────────────────────────┐            │
│  │  Docker Container                  │            │
│  │  law-scraper-cron                 │            │
│  │                                    │            │
│  │  • Cron scheduler                 │            │
│  │  • Telegram bot                   │            │
│  │  • RSS scanner                    │            │
│  └────────────────────────────────────┘            │
│                                                      │
│  data/                                              │
│  ├── keywords.json  ← управляется через Telegram   │
│  ├── rss.json                                       │
│  └── matched/                                       │
└──────────────────────────────────────────────────────┘
```

---

## 🔄 Процесс CI/CD

1. **Разработчик** делает `git push origin main`
2. **GitHub Actions** запускается автоматически
3. **Тесты** выполняются (`go test ./...`)
4. **Сборка** приложения (`go build`)
5. **SSH подключение** к серверу 77.105.133.231
6. **Копирование файлов** на сервер
7. **Docker Compose** пересоздает контейнеры
8. **Проверка** успешности деплоя
9. **Уведомление** о результате

**Время деплоя:** ~2-3 минуты

---

## 📂 Структура на сервере

```
/opt/law_scraper/
├── .env                     # Переменные окружения
├── docker-compose.yml       # Docker конфигурация
├── Dockerfile              
├── scraper/                # Исходный код
│   ├── cmd/
│   │   └── cron/main.go   # Запускается в контейнере
│   └── internal/
├── data/                   # Постоянные данные
│   ├── keywords.json       # Управляется Telegram ботом
│   ├── rss.json
│   └── matched/
│       ├── file_urls.json
│       └── file_urls.txt
└── bin/
    └── cron               # Скомпилированный бинарник
```

---

## 🛠️ Команды для управления

### Локально

```bash
# Информация о проекте
make info

# Проверка .env
make check-env

# Деплой
make deploy

# Статус
make status

# Логи
make logs-server
make logs-server-follow

# Перезапуск
make restart-server
```

### На сервере

```bash
# Подключение
ssh root@77.105.133.231

# Переход в директорию
cd /opt/law_scraper

# Статус контейнеров
docker-compose ps

# Логи
docker-compose logs -f

# Перезапуск
docker-compose restart

# Остановка
docker-compose down

# Запуск
docker-compose up -d

# Пересборка
docker-compose up -d --build
```

---

## 🐛 Устранение неполадок

### Деплой не запускается

```bash
# 1. Проверьте GitHub Secrets
# Settings → Secrets → Actions

# 2. Проверьте workflow
cat .github/workflows/deploy.yml

# 3. Проверьте логи в GitHub Actions
```

### Контейнер не запускается

```bash
# Проверьте логи
make logs-server

# На сервере
ssh root@77.105.133.231 'cd /opt/law_scraper && docker-compose logs'

# Пересоздайте контейнеры
ssh root@77.105.133.231 'cd /opt/law_scraper && docker-compose up -d --build'
```

### SSH не подключается

```bash
# Проверьте доступность
ping 77.105.133.231

# Проверьте SSH ключ
ssh -v root@77.105.133.231

# Пересоздайте SSH ключ
ssh-keygen -t rsa -b 4096
ssh-copy-id root@77.105.133.231
```

---

## 📈 Преимущества

✅ **Автоматизация** - деплой при каждом push  
✅ **Безопасность** - секреты хранятся в GitHub  
✅ **Надежность** - тесты перед деплоем  
✅ **Мониторинг** - логи и статус в один клик  
✅ **Откат** - легко откатиться к предыдущей версии  
✅ **Масштабируемость** - легко добавить новые сервера  

---

## 📖 Документация

- **Полная инструкция:** [DEPLOYMENT.md](DEPLOYMENT.md)
- **Быстрый старт:** [DEPLOYMENT_QUICKSTART.md](DEPLOYMENT_QUICKSTART.md)
- **Основная документация:** [README.md](README.md)
- **Changelog:** [CHANGELOG.md](CHANGELOG.md)

---

## 🎉 Итог

**Теперь у вас:**
1. ✅ Настроен CI/CD через GitHub Actions
2. ✅ Автоматический деплой при push в main
3. ✅ Приложение работает на 77.105.133.231
4. ✅ Удобное управление через Makefile
5. ✅ Мониторинг и логи в один клик
6. ✅ Полная документация

**Используйте:**
```bash
# Деплой
git push origin main

# Проверка
make status

# Логи
make logs-server
```

**🚀 Готово к продакшену!**

