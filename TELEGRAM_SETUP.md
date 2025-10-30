# Настройка Telegram уведомлений и крон-задачи

## 1. Создание Telegram бота

1. Найдите в Telegram бота [@BotFather](https://t.me/BotFather)
2. Отправьте команду `/newbot`
3. Следуйте инструкциям и скопируйте токен бота (выглядит как `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

## 2. Получение Chat ID

### Для личных сообщений:
1. Найдите в Telegram бота [@userinfobot](https://t.me/userinfobot)
2. Отправьте ему любое сообщение
3. Скопируйте ваш Chat ID (число)

### Для отправки в группу:
1. Добавьте вашего бота в группу
2. Отправьте сообщение в группу
3. Откройте в браузере: `https://api.telegram.org/bot<ВАШ_ТОКЕН>/getUpdates`
4. Найдите в ответе `"chat":{"id":-1234567890}` - это Chat ID группы

## 3. Настройка переменных окружения

Добавьте в файл `.env` следующие переменные:

```bash
# API настройки
API_URL=https://regulation.gov.ru/api/public/PublicProjects/GetPublicProjects

# Пути к файлам хранения
PAGES_STORAGE=./data/pages.json
MATCHED_DIR=./data/matched

# Ключевые слова для поиска (через запятую)
KEYWORDS=транспорт,концессии,закупки

# Настройки Telegram бота
TELEGRAM_BOT_TOKEN=ваш_токен_бота
TELEGRAM_CHAT_ID=ваш_chat_id

# Расписание запуска в формате cron (по умолчанию 9:00 каждый день)
# Формат: минута час день месяц день_недели
# Примеры:
#   0 9 * * *    - каждый день в 9:00
#   0 */6 * * *  - каждые 6 часов
#   0 8,20 * * * - в 8:00 и 20:00 каждый день
CRON_SCHEDULE=0 9 * * *

# Запускать ли задачу сразу при старте (true/false)
RUN_ON_START=true
```

## 4. Запуск

### Разовый запуск (для теста):
```bash
cd scraper
go run cmd/scraper/main.go
```

### Запуск крон-задачи (работает по расписанию):
```bash
cd scraper
go run cmd/cron/main.go
```

### Сборка и запуск:
```bash
cd scraper
go build -o ../bin/scraper cmd/scraper/main.go
go build -o ../bin/cron cmd/cron/main.go

# Запуск
../bin/cron
```

## 5. Формат уведомлений

Бот будет отправлять сообщения в формате:

```
🔍 Найдено совпадение

📄 Файл: [Ссылка на документ]
🔑 Ключевые слова: транспорт, концессии
```

## 6. Примеры расписаний

| Расписание | Описание |
|------------|----------|
| `0 9 * * *` | Каждый день в 9:00 |
| `0 */6 * * *` | Каждые 6 часов |
| `0 8,20 * * *` | В 8:00 и 20:00 каждый день |
| `0 9 * * 1-5` | В 9:00 по рабочим дням |
| `*/30 * * * *` | Каждые 30 минут |

## 7. Логи

Все логи выводятся в консоль. Для сохранения в файл:

```bash
../bin/cron >> logs/cron.log 2>&1
```

## 8. Systemd сервис (опционально)

Создайте файл `/etc/systemd/system/law-scraper.service`:

```ini
[Unit]
Description=Law Scraper Cron Service
After=network.target

[Service]
Type=simple
User=your_user
WorkingDirectory=/path/to/law_scraper
ExecStart=/path/to/law_scraper/bin/cron
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Запуск:
```bash
sudo systemctl enable law-scraper
sudo systemctl start law-scraper
sudo systemctl status law-scraper
```

