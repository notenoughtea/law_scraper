# 🚀 Деплой и CI/CD

Полная инструкция по развертыванию Law Scraper на продакшн сервере `77.105.133.231` и настройке автоматического деплоя через GitHub Actions.

## 📋 Содержание

- [Первоначальная настройка](#первоначальная-настройка)
- [Настройка CI/CD через GitHub Actions](#настройка-cicd-через-github-actions)
- [Ручной деплой](#ручной-деплой)
- [Управление на сервере](#управление-на-сервере)
- [Мониторинг и логи](#мониторинг-и-логи)
- [Устранение неполадок](#устранение-неполадок)

---

## Первоначальная настройка

### 1. Подготовка SSH ключа

**На локальной машине:**

```bash
# Создайте SSH ключ если его нет
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"

# Скопируйте публичный ключ на сервер
ssh-copy-id root@77.105.133.231

# Или вручную:
cat ~/.ssh/id_rsa.pub | ssh root@77.105.133.231 "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"

# Проверьте подключение
ssh root@77.105.133.231
```

### 2. Настройка сервера

**Автоматическая установка:**

```bash
# Установит Docker, Docker Compose и настроит сервер
make setup-server
```

**Или вручную на сервере:**

```bash
ssh root@77.105.133.231

# Обновление системы
apt-get update && apt-get upgrade -y

# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Установка Docker Compose
DOCKER_COMPOSE_VERSION="2.24.0"
curl -L "https://github.com/docker/compose/releases/download/v${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" \
    -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Создание директорий
mkdir -p /opt/law_scraper/data/matched
chmod -R 755 /opt/law_scraper

# Проверка
docker --version
docker-compose --version
```

### 3. Настройка .env файла

Создайте файл `.env` в корне проекта:

```bash
# API настройки
API_URL=https://regulation.gov.ru/api/public/PublicProjects/GetPublicProjects

# Пути к файлам хранения
PAGES_STORAGE=./data/pages.json
MATCHED_DIR=./data/matched

# Ключевые слова для поиска (начальные значения)
KEYWORDS=транспорт,концессии,закупки

# Telegram бот
TELEGRAM_BOT_TOKEN=ваш_токен_от_BotFather
TELEGRAM_CHAT_ID=ваш_chat_id
TELEGRAM_SEND_AS_DOCUMENT=true

# Расписание
CRON_SCHEDULE=0 9 * * *

# Запуск при старте
RUN_ON_START=false
```

---

## Настройка CI/CD через GitHub Actions

### 1. Добавление GitHub Secrets

Перейдите в настройки репозитория на GitHub:
`Settings` → `Secrets and variables` → `Actions` → `New repository secret`

Добавьте следующие секреты:

| Secret Name | Описание | Пример значения |
|------------|----------|-----------------|
| `SSH_PRIVATE_KEY` | Приватный SSH ключ | Содержимое `~/.ssh/id_rsa` |
| `SERVER_HOST` | IP адрес сервера | `77.105.133.231` |
| `SERVER_USER` | Пользователь SSH | `root` |
| `API_URL` | URL API | `https://regulation.gov.ru/api/...` |
| `KEYWORDS` | Ключевые слова | `транспорт,концессии,закупки` |
| `TELEGRAM_BOT_TOKEN` | Токен Telegram бота | `123456:ABC-DEF...` |
| `TELEGRAM_CHAT_ID` | ID чата в Telegram | `123456789` |
| `CRON_SCHEDULE` | Расписание крон | `0 9 * * *` |

#### Как получить SSH_PRIVATE_KEY:

```bash
# На локальной машине
cat ~/.ssh/id_rsa

# Скопируйте ВСЁ содержимое (включая -----BEGIN и -----END)
# И вставьте в GitHub Secret
```

### 2. Настройка workflow

Файл `.github/workflows/deploy.yml` уже создан и настроен.

**Workflow запускается автоматически при:**
- Push в ветку `main` или `master`
- Можно запустить вручную через GitHub UI

**Что делает workflow:**
1. Запускает тесты
2. Собирает приложение
3. Копирует файлы на сервер
4. Запускает Docker Compose на сервере
5. Проверяет успешность деплоя

### 3. Первый деплой через GitHub Actions

```bash
# 1. Закоммитьте изменения
git add .
git commit -m "Настройка CI/CD"
git push origin main

# 2. Проверьте статус деплоя
# Перейдите на GitHub → Actions → Deploy to Production
```

---

## Ручной деплой

### Быстрый деплой

```bash
# Деплой на сервер
make deploy
```

### Пошаговый деплой

```bash
# 1. Проверьте .env файл
make check-env

# 2. Задеплойте на сервер
make deploy

# 3. Проверьте статус
make status

# 4. Просмотрите логи
make logs-server
```

### Что делает скрипт деплоя:

1. Проверяет наличие SSH ключа и `.env` файла
2. Создает директорию на сервере
3. Копирует файлы проекта (исключая `data`, `.git`, `bin`)
4. Копирует `.env` файл
5. Запускает Docker Compose на сервере
6. Показывает статус и логи

---

## Управление на сервере

### Команды через Makefile

```bash
# Проверить статус
make status

# Просмотр логов
make logs-server

# Следить за логами в реальном времени
make logs-server-follow

# Перезапустить приложение
make restart-server
```

### Прямое подключение к серверу

```bash
# Подключение по SSH
ssh root@77.105.133.231

# Переход в директорию приложения
cd /opt/law_scraper

# Просмотр статуса контейнеров
docker-compose ps

# Просмотр логов
docker-compose logs -f

# Перезапуск
docker-compose restart

# Остановка
docker-compose down

# Запуск
docker-compose up -d

# Пересборка и запуск
docker-compose up -d --build
```

---

## Мониторинг и логи

### Просмотр логов локально

```bash
# Последние 100 строк
make logs-server

# В режиме реального времени
make logs-server-follow

# Или через SSH
ssh root@77.105.133.231 'cd /opt/law_scraper && docker-compose logs -f --tail=100'
```

### Логи на сервере

```bash
ssh root@77.105.133.231

cd /opt/law_scraper

# Логи всех сервисов
docker-compose logs -f

# Только последние 50 строк
docker-compose logs --tail=50

# Логи за последний час
docker-compose logs --since 1h

# Логи конкретного контейнера
docker-compose logs -f law-scraper-cron
```

### Проверка статуса

```bash
# Через Makefile
make status

# На сервере
ssh root@77.105.133.231 << 'EOF'
  cd /opt/law_scraper
  echo "=== Docker контейнеры ==="
  docker-compose ps
  
  echo ""
  echo "=== Использование ресурсов ==="
  docker stats --no-stream
  
  echo ""
  echo "=== Размер данных ==="
  du -sh data/
EOF
```

---

## Устранение неполадок

### Проблема: Деплой не запускается

**Решение:**

```bash
# 1. Проверьте GitHub Secrets
# Settings → Secrets → Actions

# 2. Проверьте логи workflow
# GitHub → Actions → последний workflow → логи

# 3. Проверьте SSH подключение
ssh root@77.105.133.231
```

### Проблема: Контейнер не запускается

**Решение:**

```bash
ssh root@77.105.133.231
cd /opt/law_scraper

# Проверить логи
docker-compose logs

# Проверить .env файл
cat .env

# Пересоздать контейнеры
docker-compose down
docker-compose up -d --build

# Проверить образы
docker images | grep law-scraper
```

### Проблема: Нет места на диске

**Решение:**

```bash
ssh root@77.105.133.231

# Проверить использование диска
df -h

# Очистить старые Docker образы
docker system prune -a -f

# Очистить старые логи
find /var/lib/docker/containers/ -name "*.log" -delete

# Очистить данные приложения (осторожно!)
cd /opt/law_scraper
rm -rf data/matched/*
```

### Проблема: Telegram бот не отвечает

**Решение:**

```bash
# 1. Проверьте переменные окружения
ssh root@77.105.133.231 'cd /opt/law_scraper && cat .env | grep TELEGRAM'

# 2. Проверьте логи
make logs-server-follow

# 3. Перезапустите контейнер
make restart-server

# 4. Проверьте токен бота
# Отправьте запрос к Telegram API:
curl "https://api.telegram.org/bot<YOUR_TOKEN>/getMe"
```

---

## Структура файлов на сервере

```
/opt/law_scraper/
├── .env                          # Переменные окружения
├── docker-compose.yml            # Конфигурация Docker
├── Dockerfile                    # Образ приложения
├── scraper/                      # Исходный код
│   ├── cmd/
│   │   └── cron/main.go
│   └── internal/
├── data/                         # Данные приложения
│   ├── matched/
│   │   ├── file_urls.json
│   │   └── file_urls.txt
│   ├── keywords.json             # Ключевые слова (управляются ботом)
│   ├── pages.json
│   └── rss.json
└── bin/                          # Скомпилированные бинарники
    └── cron
```

---

## Переменные окружения для деплоя

Можно использовать переменные окружения для переопределения значений по умолчанию:

```bash
# Пример с переменными окружения
SERVER_HOST=77.105.133.231 \
SERVER_USER=root \
make deploy
```

---

## Дополнительные команды

### Резервное копирование данных

```bash
# Скачать данные с сервера
scp -r root@77.105.133.231:/opt/law_scraper/data ./backup/

# Или через rsync
rsync -avz root@77.105.133.231:/opt/law_scraper/data/ ./backup/
```

### Восстановление данных

```bash
# Загрузить данные на сервер
scp -r ./backup/data root@77.105.133.231:/opt/law_scraper/

# Или через rsync
rsync -avz ./backup/data/ root@77.105.133.231:/opt/law_scraper/data/
```

### Обновление только кода (без пересборки образа)

```bash
# Быстрое обновление
ssh root@77.105.133.231 << 'EOF'
  cd /opt/law_scraper
  git pull origin main  # если используете git на сервере
  docker-compose restart
EOF
```

---

## Безопасность

### Рекомендации:

1. **SSH ключи:** Используйте SSH ключи вместо паролей
2. **Firewall:** Настройте UFW для ограничения доступа
3. **GitHub Secrets:** Храните секреты только в GitHub Secrets
4. **.env файл:** Никогда не коммитьте `.env` в репозиторий
5. **Обновления:** Регулярно обновляйте систему и Docker

### Настройка firewall (опционально):

```bash
ssh root@77.105.133.231

# Разрешить SSH
ufw allow 22/tcp

# Включить firewall
ufw --force enable

# Проверить статус
ufw status
```

---

## Автоматизация

### Настройка webhook (опционально)

Можно настроить webhook для автоматического деплоя при push в GitHub без использования GitHub Actions:

```bash
# На сервере создайте webhook listener
# См. документацию: https://github.com/adnanh/webhook
```

---

## Мониторинг производительности

### Проверка ресурсов

```bash
ssh root@77.105.133.231

# Использование CPU/Memory
docker stats

# Размер образов
docker images

# Размер данных
du -sh /opt/law_scraper/data/

# Логи Docker
journalctl -u docker --since "1 hour ago"
```

---

## Полезные ссылки

- [Docker Documentation](https://docs.docker.com/)
- [GitHub Actions Documentation](https://docs.github.com/actions)
- [Docker Compose Reference](https://docs.docker.com/compose/compose-file/)

---

**✅ Теперь у вас настроен полный CI/CD pipeline!**

При каждом push в `main` приложение автоматически задеплоится на сервер `77.105.133.231`.

