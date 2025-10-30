# 🚀 Быстрый старт деплоя на 77.105.133.231

## За 5 минут

### 1. Подготовка SSH

```bash
# Создайте SSH ключ (если его нет)
ssh-keygen -t rsa -b 4096

# Скопируйте ключ на сервер
ssh-copy-id root@77.105.133.231

# Проверьте подключение
ssh root@77.105.133.231
```

### 2. Настройка сервера (первый раз)

```bash
# Установит Docker и Docker Compose автоматически
make setup-server
```

### 3. Деплой приложения

```bash
# Убедитесь что файл .env настроен
make check-env

# Задеплойте на сервер
make deploy
```

**Готово!** Приложение работает на сервере.

---

## Проверка работы

```bash
# Проверить статус
make status

# Посмотреть логи
make logs-server
```

---

## Настройка автоматического деплоя через GitHub Actions

### 1. Добавьте GitHub Secrets

Перейдите: `GitHub Repo` → `Settings` → `Secrets and variables` → `Actions`

Добавьте секреты:

```bash
# SSH_PRIVATE_KEY
cat ~/.ssh/id_rsa
# Скопируйте ВСЁ содержимое и добавьте как секрет

# SERVER_HOST
77.105.133.231

# SERVER_USER
root

# TELEGRAM_BOT_TOKEN
ваш_токен_от_BotFather

# TELEGRAM_CHAT_ID
ваш_chat_id

# KEYWORDS
транспорт,концессии,закупки

# API_URL
https://regulation.gov.ru/api/public/PublicProjects/GetPublicProjects

# CRON_SCHEDULE
0 9 * * *
```

### 2. Push в GitHub

```bash
git add .
git commit -m "Setup deployment"
git push origin main
```

**Готово!** Теперь при каждом push приложение автоматически деплоится.

---

## Управление

```bash
# Статус на сервере
make status

# Логи
make logs-server

# Перезапуск
make restart-server

# Подключение к серверу
ssh root@77.105.133.231
```

---

## Если что-то пошло не так

### Проблема: SSH не подключается

```bash
# Проверьте доступность сервера
ping 77.105.133.231

# Проверьте SSH
ssh -v root@77.105.133.231
```

### Проблема: Деплой не работает

```bash
# Проверьте .env файл
cat .env

# Проверьте логи на сервере
ssh root@77.105.133.231 'cd /opt/law_scraper && docker-compose logs'
```

### Проблема: GitHub Actions не запускается

```bash
# Проверьте что добавлены все Secrets
# GitHub → Settings → Secrets → Actions

# Проверьте workflow файл
cat .github/workflows/deploy.yml
```

---

## Полная документация

📖 Смотрите [DEPLOYMENT.md](DEPLOYMENT.md) для подробной информации.

---

**✅ Деплой настроен! Наслаждайтесь автоматизацией!**

