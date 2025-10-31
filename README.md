# Law Scraper - Мониторинг нормативных актов

Автоматический сканер нормативных актов с https://regulation.gov.ru с поиском по ключевым словам и отправкой уведомлений в Telegram.

## 🚀 Быстрый старт

### 1. Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```bash
# API настройки
API_URL=https://regulation.gov.ru/api/public/PublicProjects/GetPublicProjects

# Пути к файлам хранения
PAGES_STORAGE=./data/pages.json
MATCHED_DIR=./data/matched

# Ключевые слова для поиска (через запятую, регистр не важен)
# ВНИМАНИЕ: Если используете интерактивного бота, ключевые слова 
# можно менять через Telegram командами, и они сохранятся в data/keywords.json
KEYWORDS=транспорт,концессии,закупки

# Telegram бот (см. TELEGRAM_SETUP.md)
TELEGRAM_BOT_TOKEN=ваш_токен_от_BotFather
TELEGRAM_CHAT_ID=ваш_chat_id

# Расписание (формат cron)
CRON_SCHEDULE=0 9 * * *

# Количество параллельных воркеров для обработки файлов (по умолчанию 3)
# Для слабого сервера: 2-3, для мощного: 5-8
MAX_WORKERS=3

# Запуск при старте
RUN_ON_START=true
```

📖 **Подробные инструкции:**
- Настройка Telegram бота: [TELEGRAM_SETUP.md](TELEGRAM_SETUP.md)
- Команды управления через бота: [TELEGRAM_BOT_COMMANDS.md](TELEGRAM_BOT_COMMANDS.md)
- Деплой на сервер и CI/CD: [DEPLOYMENT.md](DEPLOYMENT.md)
- Быстрый старт деплоя: [DEPLOYMENT_QUICKSTART.md](DEPLOYMENT_QUICKSTART.md)

### 2. Установка зависимостей

```bash
cd scraper
go mod download
```

### 3. Запуск

#### Вариант A: Разовый запуск (для теста)
```bash
cd scraper
go run cmd/scraper/main.go
```

#### Вариант B: Запуск по расписанию
```bash
cd scraper
go run cmd/cron/main.go
```

#### Вариант C: Сборка и запуск
```bash
cd scraper
go build -o ../bin/cron cmd/cron/main.go
cd ..
./bin/cron
```

#### Вариант D: Docker Compose
```bash
docker-compose up -d
docker-compose logs -f  # просмотр логов
```

#### Вариант E: Деплой на продакшн сервер (77.105.133.231)
```bash
# Первый раз: настройка сервера
make setup-server

# Деплой приложения
make deploy

# Проверка статуса
make status

# Просмотр логов
make logs-server
```

📖 **Полная инструкция по деплою**: [DEPLOYMENT_QUICKSTART.md](DEPLOYMENT_QUICKSTART.md)

#### Вариант E: Интерактивный Telegram бот
```bash
cd scraper
go run cmd/bot/main.go
```

Или с Docker:
```bash
docker-compose up -d law-scraper-bot
docker-compose logs -f law-scraper-bot
```

## 📋 Как это работает

1. **Сканирование**: Скрапер загружает RSS-ленту с regulation.gov.ru
2. **Поиск**: Ищет ключевые слова в документах каждого проекта
3. **Сохранение**: Сохраняет найденные URL в `data/matched/file_urls.json`
4. **Уведомления**: Отправляет каждую найденную ссылку в Telegram
5. **Управление**: Интерактивный Telegram бот для изменения ключевых слов на лету

## 📁 Структура проекта

```
law_scraper/
├── scraper/
│   ├── cmd/
│   │   ├── scraper/main.go    # Разовый запуск
│   │   └── cron/main.go       # Запуск по расписанию
│   ├── internal/
│   │   ├── clients/           # HTTP клиенты (RSS, API, Telegram)
│   │   ├── config/            # Конфигурация из .env
│   │   ├── service/           # Бизнес-логика (scanner, notifier)
│   │   ├── repository/        # Работа с файлами
│   │   └── logger/            # Логирование
│   ├── go.mod
│   └── go.sum
├── data/
│   └── matched/
│       └── file_urls.txt      # Результаты поиска
├── .env                       # Настройки (не в git)
├── docker-compose.yml
├── Dockerfile
├── TELEGRAM_SETUP.md          # Инструкция по Telegram
└── README.md
```

## 🔧 Конфигурация

### Формат расписания (CRON_SCHEDULE)

| Расписание | Описание |
|------------|----------|
| `0 9 * * *` | Каждый день в 9:00 |
| `0 */6 * * *` | Каждые 6 часов |
| `0 8,20 * * *` | В 8:00 и 20:00 |
| `0 9 * * 1-5` | В 9:00 по рабочим дням |
| `*/30 * * * *` | Каждые 30 минут |

### Ключевые слова

- Указываются через запятую в `.env` или через Telegram бота
- Регистр не важен (поиск в нижнем регистре)
- Примеры: `транспорт`, `концессии`, `государственные закупки`
- **Управление через Telegram бота** (рекомендуется):
  - `/keywords` - показать текущие слова
  - `/set_keywords слово1,слово2` - установить новые
  - `/add_keyword слово` - добавить одно слово
  - `/remove_keyword слово` - удалить слово
- Слова сохраняются в `data/keywords.json` и применяются сразу

## 🤖 Управление через Telegram бота

Для удобного управления ключевыми словами без редактирования файлов используйте интерактивного Telegram бота.

### Запуск бота

Telegram бот **автоматически запускается** вместе с крон-планировщиком и работает параллельно со сканированием.

**Локально:**
```bash
cd scraper
go run cmd/cron/main.go
```

**Docker:**
```bash
# Запуск крон-планировщика с интегрированным ботом
docker-compose up -d

# Просмотр логов
docker-compose logs -f law-scraper-cron
```

**Примечания:**
- Бот запускается автоматически в фоновом режиме
- Не требуется отдельный процесс для бота
- Бот работает 24/7 и принимает команды в любое время

### Доступные команды

| Команда | Описание | Пример |
|---------|----------|--------|
| `/start` | Приветствие и информация о боте | `/start` |
| `/help` | Показать справку по командам | `/help` |
| `/keywords` | Показать текущие ключевые слова | `/keywords` |
| `/set_keywords` | Установить новый список слов | `/set_keywords транспорт,образование` |
| `/add_keyword` | Добавить одно ключевое слово | `/add_keyword экология` |
| `/remove_keyword` | Удалить ключевое слово | `/remove_keyword транспорт` |

### Примеры использования

**Установка новых слов:**
```
/set_keywords транспорт,образование,здравоохранение,экология
```

**Добавление слова:**
```
/add_keyword строительство
```

**Удаление слова:**
```
/remove_keyword транспорт
```

**Просмотр текущих слов:**
```
/keywords
```

### Важно
- Изменения применяются **мгновенно** и сохраняются в `data/keywords.json`
- Ключевые слова из файла имеют приоритет над `.env`
- Регистр не важен, все слова автоматически приводятся к нижнему регистру
- Бот работает 24/7 и принимает команды в любое время

## 📨 Формат уведомлений

Telegram-бот отправляет сообщения в формате:

```
🔍 Найдено совпадение

📄 Файл: [Ссылка на документ]
🔑 Ключевые слова: транспорт, концессии
```

## 🛠 Разработка

### Сборка всех бинарников
```bash
cd scraper
go build -o ../bin/scraper cmd/scraper/main.go
go build -o ../bin/cron cmd/cron/main.go
```

### Запуск тестов
```bash
cd scraper
go test ./...
```

### Просмотр логов
```bash
# При запуске через docker-compose
docker-compose logs -f

# При ручном запуске с сохранением логов
./bin/cron 2>&1 | tee logs/cron.log
```

## 🐳 Docker

### Сборка образа
```bash
docker build -t law-scraper .
```

### Запуск
```bash
docker-compose up -d
```

### Остановка
```bash
docker-compose down
```

## 📝 Системный сервис (Linux)

Для автоматического запуска при загрузке системы см. раздел 8 в [TELEGRAM_SETUP.md](TELEGRAM_SETUP.md)

## 🔍 Поддерживаемые форматы документов

- PDF
- DOCX (Microsoft Word)
- TXT
- DOC (старый формат Word)

## ⚙️ Требования

- Go 1.24+
- Доступ к интернету
- Telegram бот токен

## 📄 Лицензия

MIT

## 🤝 Контрибуция

Pull requests приветствуются!

