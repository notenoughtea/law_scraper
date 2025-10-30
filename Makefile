.PHONY: help build run-scraper run-cron run-bot test-telegram docker-build docker-up docker-down clean

help: ## Показать эту справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать все бинарники
	@echo "Сборка бинарников..."
	@mkdir -p bin
	@cd scraper && go build -o ../bin/scraper cmd/scraper/main.go
	@cd scraper && go build -o ../bin/cron cmd/cron/main.go
	@cd scraper && go build -o ../bin/bot cmd/bot/main.go
	@cd scraper && go build -o ../bin/test-telegram cmd/test-telegram/main.go
	@echo "✅ Сборка завершена! Бинарники в папке bin/"

run-scraper: ## Запустить разовое сканирование
	@echo "Запуск сканирования..."
	@cd scraper && go run cmd/scraper/main.go

run-cron: ## Запустить крон-задачу
	@echo "Запуск крон-планировщика..."
	@cd scraper && go run cmd/cron/main.go

run-bot: ## Запустить интерактивного Telegram бота
	@echo "Запуск Telegram бота..."
	@cd scraper && go run cmd/bot/main.go

test-telegram: ## Проверить настройки Telegram
	@echo "Отправка тестового сообщения в Telegram..."
	@cd scraper && go run cmd/test-telegram/main.go

docker-build: ## Собрать Docker образ
	@echo "Сборка Docker образа..."
	@docker build -t law-scraper .

docker-up: ## Запустить через Docker Compose
	@echo "Запуск Docker Compose..."
	@docker-compose up -d
	@echo "✅ Сервис запущен!"
	@echo "Просмотр логов: make docker-logs"

docker-down: ## Остановить Docker Compose
	@echo "Остановка Docker Compose..."
	@docker-compose down

docker-logs: ## Показать логи Docker Compose
	@docker-compose logs -f

docker-logs-bot: ## Показать логи Telegram бота
	@docker-compose logs -f law-scraper-bot

docker-logs-cron: ## Показать логи крон-задачи
	@docker-compose logs -f law-scraper-cron

docker-restart: docker-down docker-up ## Перезапустить Docker Compose

clean: ## Очистить собранные файлы
	@echo "Очистка..."
	@rm -rf bin/
	@rm -rf data/matched/*
	@echo "✅ Очистка завершена"

deps: ## Установить зависимости
	@echo "Установка зависимостей..."
	@cd scraper && go mod download
	@echo "✅ Зависимости установлены"

check-env: ## Проверить .env файл
	@if [ ! -f .env ]; then \
		echo "❌ Файл .env не найден!"; \
		echo "Создайте .env файл на основе TELEGRAM_SETUP.md"; \
		exit 1; \
	else \
		echo "✅ Файл .env найден"; \
	fi

# =============================================================================
# Деплой команды
# =============================================================================

setup-server: ## Первоначальная настройка сервера (77.105.133.231)
	@echo "🔧 Настройка сервера..."
	@chmod +x deployment/scripts/setup-server.sh
	@./deployment/scripts/setup-server.sh

deploy: ## Деплой на продакшн сервер
	@echo "🚀 Деплой на сервер..."
	@chmod +x deployment/scripts/deploy.sh
	@./deployment/scripts/deploy.sh

status: ## Проверить статус на сервере
	@echo "📊 Проверка статуса..."
	@chmod +x deployment/scripts/status.sh
	@./deployment/scripts/status.sh

logs-server: ## Показать логи с сервера
	@echo "📋 Просмотр логов..."
	@chmod +x deployment/scripts/logs.sh
	@./deployment/scripts/logs.sh

logs-server-follow: ## Следить за логами на сервере
	@echo "📋 Просмотр логов (following)..."
	@chmod +x deployment/scripts/logs.sh
	@./deployment/scripts/logs.sh --follow

restart-server: ## Перезапустить приложение на сервере
	@echo "🔄 Перезапуск..."
	@chmod +x deployment/scripts/restart.sh
	@./deployment/scripts/restart.sh

# =============================================================================
# Информация
# =============================================================================

info: ## Показать информацию о проекте
	@echo "╔════════════════════════════════════════════════════════════╗"
	@echo "║           Law Scraper - Мониторинг НПА                     ║"
	@echo "╚════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "📍 Продакшн сервер: 77.105.133.231"
	@echo "📁 Директория на сервере: /opt/law_scraper"
	@echo ""
	@echo "Основные команды:"
	@echo "  make deploy          - Деплой на сервер"
	@echo "  make status          - Статус на сервере"
	@echo "  make logs-server     - Логи с сервера"
	@echo "  make restart-server  - Перезапуск на сервере"
	@echo ""
	@echo "Документация:"
	@echo "  DEPLOYMENT.md        - Инструкция по деплою"
	@echo "  TELEGRAM_SETUP.md    - Настройка Telegram"
	@echo "  README.md            - Основная документация"

