#!/bin/bash

# Скрипт проверки статуса на сервере
# Использование: ./status.sh

SERVER_HOST="${SERVER_HOST:-77.105.133.231}"
SERVER_USER="${SERVER_USER:-root}"
APP_DIR="/opt/law_scraper"

echo "📊 Проверка статуса на $SERVER_HOST..."
echo ""

ssh $SERVER_USER@$SERVER_HOST << ENDSSH
    cd $APP_DIR
    
    echo "🐳 Статус Docker контейнеров:"
    docker-compose ps
    
    echo ""
    echo "💾 Использование диска:"
    df -h | grep -E '(Filesystem|/$|/opt)'
    
    echo ""
    echo "📁 Размер директории приложения:"
    du -sh $APP_DIR
    du -sh $APP_DIR/data
    
    echo ""
    echo "📋 Последние 10 строк логов:"
    docker-compose logs --tail=10
ENDSSH

echo ""
echo "✅ Проверка завершена"

