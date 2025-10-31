.PHONY: help build run-scraper run-cron run-bot test-telegram docker-build docker-up docker-down clean check-commit

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

add-ssh-key: ## Добавить публичный SSH ключ на сервер
	@echo "📤 Добавление публичного SSH ключа на сервер..."
	@chmod +x deployment/scripts/add-ssh-key.sh
	@./deployment/scripts/add-ssh-key.sh

update-deployment-info: ## Создать и отправить .deployment_info на сервер
	@echo "📦 Создание файла .deployment_info..."
	@if ! git rev-parse --git-dir > /dev/null 2>&1; then \
		echo "❌ Не git репозиторий"; \
		exit 1; \
	fi
	@COMMIT_HASH=$$(git rev-parse HEAD); \
	COMMIT_AUTHOR=$$(git log -1 --pretty=format:'%an <%ae>' | sed 's/"/\\"/g'); \
	COMMIT_MESSAGE=$$(git log -1 --pretty=format:'%s' | sed 's/"/\\"/g' | sed "s/'/\\'/g"); \
	DEPLOY_DATE=$$(date -u +'%Y-%m-%d %H:%M:%S UTC'); \
	BRANCH=$$(git branch --show-current || git rev-parse --abbrev-ref HEAD | sed 's/"/\\"/g'); \
	echo "COMMIT_HASH=\"$$COMMIT_HASH\"" > .deployment_info; \
	echo "COMMIT_AUTHOR=\"$$COMMIT_AUTHOR\"" >> .deployment_info; \
	echo "COMMIT_MESSAGE=\"$$COMMIT_MESSAGE\"" >> .deployment_info; \
	echo "DEPLOY_DATE=\"$$DEPLOY_DATE\"" >> .deployment_info; \
	echo "BRANCH=\"$$BRANCH\"" >> .deployment_info; \
	echo "WORKFLOW_RUN=\"manual\"" >> .deployment_info; \
	echo "✅ Файл .deployment_info создан:"; \
	cat .deployment_info; \
	echo ""; \
	echo "📤 Копируем на сервер..."; \
	scp -o StrictHostKeyChecking=no .deployment_info root@77.105.133.231:/opt/law_scraper/.deployment_info && \
	echo "✅ Файл успешно скопирован на сервер!" || \
	(echo "❌ Ошибка копирования"; exit 1)

check-commit: ## Проверить последний ли коммит на сервере
	@echo "🔍 Проверка последнего коммита на сервере..."
	@echo ""
	@echo "📋 Локальный репозиторий:"
	@if git rev-parse --git-dir > /dev/null 2>&1; then \
		git log -1 --pretty=format:"  Хеш: %H%n  Автор: %an <%ae>%n  Дата: %ad%n  Сообщение: %s" --date=format:"%Y-%m-%d %H:%M:%S"; \
	else \
		echo "  ⚠️  Не git репозиторий"; \
	fi
	@echo ""
	@echo "📋 На сервере (77.105.133.231):"
	@SSH_OUTPUT=$$(ssh -o StrictHostKeyChecking=no -o ConnectTimeout=10 -o BatchMode=yes root@77.105.133.231 'cd /opt/law_scraper 2>/dev/null && if [ -d .git ]; then echo "GIT_FOUND"; git log -1 --pretty=format:"  Хеш: %H%n  Автор: %an <%ae>%n  Дата: %ad%n  Сообщение: %s" --date=format:"%Y-%m-%d %H:%M:%S" 2>/dev/null; elif [ -d scraper ]; then echo "NO_GIT"; echo "  📁 Git репозиторий не найден на сервере"; if [ -f .deployment_info ]; then echo "  📦 Информация о деплое:"; COMMIT_HASH_VAL=$$(grep "^COMMIT_HASH=" .deployment_info 2>/dev/null | cut -d"=" -f2- | sed "s/^\"//; s/\"\$$//"); COMMIT_AUTHOR_VAL=$$(grep "^COMMIT_AUTHOR=" .deployment_info 2>/dev/null | cut -d"=" -f2- | sed "s/^\"//; s/\"\$$//"); DEPLOY_DATE_VAL=$$(grep "^DEPLOY_DATE=" .deployment_info 2>/dev/null | cut -d"=" -f2- | sed "s/^\"//; s/\"\$$//"); COMMIT_MESSAGE_VAL=$$(grep "^COMMIT_MESSAGE=" .deployment_info 2>/dev/null | cut -d"=" -f2- | sed "s/^\"//; s/\"\$$//"); BRANCH_VAL=$$(grep "^BRANCH=" .deployment_info 2>/dev/null | cut -d"=" -f2- | sed "s/^\"//; s/\"\$$//"); [ -n "$$COMMIT_HASH_VAL" ] && echo "    Хеш коммита: $$COMMIT_HASH_VAL"; [ -n "$$COMMIT_AUTHOR_VAL" ] && echo "    Автор: $$COMMIT_AUTHOR_VAL"; [ -n "$$DEPLOY_DATE_VAL" ] && echo "    Дата деплоя: $$DEPLOY_DATE_VAL"; [ -n "$$COMMIT_MESSAGE_VAL" ] && echo "    Сообщение: $$COMMIT_MESSAGE_VAL"; [ -n "$$BRANCH_VAL" ] && echo "    Ветка: $$BRANCH_VAL"; else echo "  ⚠️  Файл .deployment_info не найден"; fi; echo "  📋 Проверка версии через Docker:"; if command -v docker-compose > /dev/null 2>&1; then docker-compose ps 2>/dev/null | grep -q "Up" && echo "    ✅ Контейнеры запущены" || echo "    ⚠️  Контейнеры не запущены"; fi; else echo "NO_DIR"; echo "  ❌ Директория /opt/law_scraper не найдена"; fi' 2>&1); \
	SSH_EXIT=$$?; \
	if [ $$SSH_EXIT -eq 0 ]; then \
		echo "$$SSH_OUTPUT"; \
	elif echo "$$SSH_OUTPUT" | grep -q "Permission denied"; then \
		echo "  ❌ Ошибка: Permission denied"; \
		echo "  💡 Проверьте SSH ключ:"; \
		echo "     ssh-keygen -t rsa -b 4096"; \
		echo "     ssh-copy-id root@77.105.133.231"; \
	elif echo "$$SSH_OUTPUT" | grep -q "Connection refused\|Connection timed out"; then \
		echo "  ❌ Ошибка: Не удалось подключиться к серверу"; \
		echo "  💡 Проверьте:"; \
		echo "     - Доступность сервера: ping 77.105.133.231"; \
		echo "     - SSH сервис на сервере: ssh root@77.105.133.231"; \
	else \
		echo "  ❌ Ошибка подключения"; \
		echo "  💡 Ошибка: $$SSH_OUTPUT"; \
	fi
	@echo ""
	@echo "🔍 Сравнение:"
	@LOCAL_HASH=$$(git rev-parse HEAD 2>/dev/null); \
	if [ -n "$$LOCAL_HASH" ]; then \
		chmod +x deployment/scripts/get-commit-hash.sh 2>/dev/null; \
		REMOTE_HASH=$$(./deployment/scripts/get-commit-hash.sh 2>/dev/null); \
		if [ -z "$$REMOTE_HASH" ]; then \
			echo "  ⚠️  Не удалось получить коммит с сервера"; \
			echo "     Локальный: $$LOCAL_HASH"; \
			echo ""; \
			echo "  💡 Возможные причины:"; \
			echo "     - На сервере нет git репозитория и файла .deployment_info"; \
			echo "     - Проблемы с SSH подключением"; \
			echo "  💡 Для обновления: make deploy"; \
		elif [ "$$REMOTE_HASH" = "$$LOCAL_HASH" ]; then \
			echo "  ✅ Коммиты совпадают - на сервере последняя версия!"; \
		else \
			echo "  ⚠️  Коммиты отличаются:"; \
			echo "     Локальный:  $$LOCAL_HASH"; \
			echo "     На сервере: $$REMOTE_HASH"; \
			echo ""; \
			echo "  💡 Для обновления: make deploy"; \
		fi; \
	else \
		echo "  ⚠️  Не git репозиторий локально"; \
	fi

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
	@echo "  make check-commit    - Проверить последний коммит на сервере"
	@echo "  make logs-server     - Логи с сервера"
	@echo "  make restart-server  - Перезапуск на сервере"
	@echo ""
	@echo "Документация:"
	@echo "  DEPLOYMENT.md        - Инструкция по деплою"
	@echo "  TELEGRAM_SETUP.md    - Настройка Telegram"
	@echo "  README.md            - Основная документация"

