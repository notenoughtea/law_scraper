#!/bin/bash

# Скрипт перезапуска приложения на сервере
# Использование: ./restart.sh

SERVER_HOST="${SERVER_HOST:-77.105.133.231}"
SERVER_USER="${SERVER_USER:-root}"
APP_DIR="/opt/law_scraper"

echo "🔄 Перезапуск приложения на $SERVER_HOST..."

ssh $SERVER_USER@$SERVER_HOST << ENDSSH
    cd $APP_DIR
    
    echo "🛑 Остановка контейнеров..."
    docker-compose down
    
    echo "🚀 Запуск контейнеров..."
    docker-compose up -d
    
    echo ""
    echo "✅ Статус после перезапуска:"
    docker-compose ps
    
    echo ""
    echo "📋 Последние логи:"
    docker-compose logs --tail=20
ENDSSH

echo ""
echo "✅ Перезапуск завершен"

