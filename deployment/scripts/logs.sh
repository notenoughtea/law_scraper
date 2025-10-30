#!/bin/bash

# Скрипт для просмотра логов на сервере
# Использование: ./logs.sh [--follow]

SERVER_HOST="${SERVER_HOST:-77.105.133.231}"
SERVER_USER="${SERVER_USER:-root}"
APP_DIR="/opt/law_scraper"

FOLLOW_FLAG=""
if [ "$1" = "--follow" ] || [ "$1" = "-f" ]; then
    FOLLOW_FLAG="-f"
fi

echo "📋 Подключение к логам на $SERVER_HOST..."

ssh -t $SERVER_USER@$SERVER_HOST << ENDSSH
    cd $APP_DIR
    docker-compose logs $FOLLOW_FLAG --tail=100
ENDSSH

