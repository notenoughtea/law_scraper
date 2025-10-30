#!/bin/bash
set -e

# Скрипт первоначальной настройки сервера
# Использование: ./setup-server.sh

SERVER_HOST="${SERVER_HOST:-77.105.133.231}"
SERVER_USER="${SERVER_USER:-root}"

echo "🔧 Настройка сервера $SERVER_HOST..."

ssh $SERVER_USER@$SERVER_HOST << 'ENDSSH'
    set -e
    
    echo "📦 Обновление системы..."
    apt-get update
    apt-get upgrade -y
    
    echo "🐳 Установка Docker..."
    if ! command -v docker &> /dev/null; then
        # Установка Docker
        curl -fsSL https://get.docker.com -o get-docker.sh
        sh get-docker.sh
        rm get-docker.sh
        
        # Добавление пользователя в группу docker
        usermod -aG docker $USER || true
    else
        echo "✓ Docker уже установлен"
    fi
    
    echo "🐙 Установка Docker Compose..."
    if ! command -v docker-compose &> /dev/null; then
        # Установка Docker Compose
        DOCKER_COMPOSE_VERSION="2.24.0"
        curl -L "https://github.com/docker/compose/releases/download/v${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" \
            -o /usr/local/bin/docker-compose
        chmod +x /usr/local/bin/docker-compose
    else
        echo "✓ Docker Compose уже установлен"
    fi
    
    echo "📁 Создание директорий..."
    mkdir -p /opt/law_scraper/data/matched
    chmod -R 755 /opt/law_scraper
    
    echo "🔥 Настройка firewall (UFW)..."
    if command -v ufw &> /dev/null; then
        ufw --force enable
        ufw allow 22/tcp  # SSH
        ufw status
    fi
    
    echo "✅ Сервер настроен успешно!"
    echo ""
    echo "Установленные версии:"
    docker --version
    docker-compose --version
ENDSSH

echo ""
echo "✅ Настройка сервера завершена!"
echo ""
echo "Следующие шаги:"
echo "1. Настройте .env файл локально"
echo "2. Запустите деплой: ./deployment/scripts/deploy.sh"

