.PHONY: help build run-scraper run-cron test-telegram docker-build docker-up docker-down clean

help: ## Показать эту справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать все бинарники
	@echo "Сборка бинарников..."
	@mkdir -p bin
	@cd scraper && go build -o ../bin/scraper cmd/scraper/main.go
	@cd scraper && go build -o ../bin/cron cmd/cron/main.go
	@cd scraper && go build -o ../bin/test-telegram cmd/test-telegram/main.go
	@echo "✅ Сборка завершена! Бинарники в папке bin/"

run-scraper: ## Запустить разовое сканирование
	@echo "Запуск сканирования..."
	@cd scraper && go run cmd/scraper/main.go

run-cron: ## Запустить крон-задачу
	@echo "Запуск крон-планировщика..."
	@cd scraper && go run cmd/cron/main.go

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

