#!/bin/bash
# Скрипт для добавления публичного SSH ключа на сервер

SERVER_HOST="77.105.133.231"
SERVER_USER="root"

echo "📤 Добавление публичного SSH ключа на сервер..."
echo ""

# Проверяем наличие публичного ключа
if [ ! -f ~/.ssh/id_rsa.pub ]; then
    echo "❌ Файл ~/.ssh/id_rsa.pub не найден!"
    exit 1
fi

PUBLIC_KEY=$(cat ~/.ssh/id_rsa.pub)

echo "🔑 Публичный ключ:"
echo "$PUBLIC_KEY"
echo ""

# Пытаемся добавить через ssh-copy-id
if ssh-copy-id -i ~/.ssh/id_rsa.pub "$SERVER_USER@$SERVER_HOST" 2>/dev/null; then
    echo "✅ Ключ успешно добавлен на сервер через ssh-copy-id!"
    exit 0
fi

echo "⚠️  ssh-copy-id не сработал, пробуем вручную..."
echo ""

# Пытаемся добавить вручную
if ssh "$SERVER_USER@$SERVER_HOST" "mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '$PUBLIC_KEY' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys" 2>/dev/null; then
    echo "✅ Ключ успешно добавлен на сервер!"
    exit 0
fi

echo "❌ Не удалось добавить ключ автоматически"
echo ""
echo "💡 Добавьте ключ вручную:"
echo "   1. Подключитесь к серверу: ssh $SERVER_USER@$SERVER_HOST"
echo "   2. Выполните на сервере:"
echo "      mkdir -p ~/.ssh"
echo "      chmod 700 ~/.ssh"
echo "      echo '$PUBLIC_KEY' >> ~/.ssh/authorized_keys"
echo "      chmod 600 ~/.ssh/authorized_keys"
echo ""
echo "📋 Или скопируйте этот ключ и добавьте на сервер:"
echo "$PUBLIC_KEY"

exit 1

