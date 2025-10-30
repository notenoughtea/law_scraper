#!/bin/bash
set -e

# Скрипт деплоя на продакшн сервер
# Использование: ./deploy.sh

SERVER_HOST="${SERVER_HOST:-77.105.133.231}"
SERVER_USER="${SERVER_USER:-root}"
APP_DIR="/opt/law_scraper"

echo "🚀 Начало деплоя на $SERVER_HOST..."

# Проверка SSH ключа
if [ ! -f ~/.ssh/id_rsa ]; then
    echo "❌ SSH ключ не найден в ~/.ssh/id_rsa"
    echo "Создайте SSH ключ: ssh-keygen -t rsa -b 4096"
    exit 1
fi

# Проверка .env файла
if [ ! -f .env ]; then
    echo "❌ Файл .env не найден"
    echo "Создайте файл .env на основе .env.example"
    exit 1
fi

echo "📦 Создание директории на сервере..."
ssh $SERVER_USER@$SERVER_HOST "mkdir -p $APP_DIR"

echo "📤 Копирование файлов на сервер..."
rsync -avz --exclude 'data' \
           --exclude '.git' \
           --exclude 'bin' \
           --exclude '.env' \
           ./ $SERVER_USER@$SERVER_HOST:$APP_DIR/

echo "📤 Копирование .env файла..."
scp .env $SERVER_USER@$SERVER_HOST:$APP_DIR/.env

echo "🐳 Запуск Docker Compose на сервере..."
ssh $SERVER_USER@$SERVER_HOST << 'ENDSSH'
    cd /opt/law_scraper
    
    # Останавливаем старые контейнеры
    docker-compose down || true
    
    # Удаляем старые образы (опционально)
    # docker-compose rm -f || true
    
    # Собираем новый образ
    docker-compose build
    
    # Запускаем контейнеры
    docker-compose up -d
    
    # Показываем статус
    echo ""
    echo "✅ Статус контейнеров:"
    docker-compose ps
    
    echo ""
    echo "📋 Последние логи:"
    docker-compose logs --tail=20
ENDSSH

echo ""
echo "✅ Деплой завершен успешно!"
echo "📊 Проверьте логи: ssh $SERVER_USER@$SERVER_HOST 'cd $APP_DIR && docker-compose logs -f'"

