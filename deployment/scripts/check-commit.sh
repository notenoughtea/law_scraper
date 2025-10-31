#!/bin/bash
# Скрипт для проверки последнего коммита на сервере

SERVER_HOST="77.105.133.231"
SERVER_USER="root"
SERVER_DIR="/opt/law_scraper"

echo "🔍 Проверка последнего коммита на сервере..."
echo ""

# Получаем локальный коммит
echo "📋 Локальный репозиторий:"
LOCAL_HASH=$(git rev-parse HEAD 2>/dev/null)
if [ -n "$LOCAL_HASH" ]; then
    git log -1 --pretty=format:"  Хеш: %H%n  Автор: %an <%ae>%n  Дата: %ad%n  Сообщение: %s" --date=format:"%Y-%m-%d %H:%M:%S"
    echo ""
else
    echo "  ⚠️  Не git репозиторий"
    echo ""
fi

# Получаем коммит с сервера
echo "📋 На сервере ($SERVER_HOST):"
if ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 "$SERVER_USER@$SERVER_HOST" \
    "test -d $SERVER_DIR/.git" 2>/dev/null; then
    
    REMOTE_HASH=$(ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 \
        "$SERVER_USER@$SERVER_HOST" \
        "cd $SERVER_DIR && git rev-parse HEAD 2>&1 | grep -v 'fatal:' | head -1" 2>/dev/null | grep -E '^[a-f0-9]{40}$' || echo "")
    
    if [ -n "$REMOTE_HASH" ]; then
        ssh -o StrictHostKeyChecking=no "$SERVER_USER@$SERVER_HOST" \
            "cd $SERVER_DIR && git log -1 --pretty=format:'  Хеш: %H%n  Автор: %an <%ae>%n  Дата: %ad%n  Сообщение: %s' --date=format:'%Y-%m-%d %H:%M:%S' 2>&1 | grep -v 'fatal:'" 2>/dev/null || echo "  ⚠️  Не удалось получить информацию из git"
        echo ""
    else
        echo "  ⚠️  Не удалось получить коммит"
    fi
else
    echo "  Git репозиторий не найден на сервере"
    echo "  📅 Дата последнего изменения файлов:"
    ssh -o StrictHostKeyChecking=no "$SERVER_USER@$SERVER_HOST" \
        "ls -ltd $SERVER_DIR/scraper/*.go 2>/dev/null | head -1 | awk '{print \"    \" \$6, \$7, \$8, \$9}' || echo '    Не удалось определить'" 2>/dev/null || echo "    ❌ Не удалось подключиться"
fi

echo ""
echo "🔍 Сравнение:"

if [ -n "$LOCAL_HASH" ] && [ -n "$REMOTE_HASH" ]; then
    if [ "$LOCAL_HASH" = "$REMOTE_HASH" ]; then
        echo "  ✅ Коммиты совпадают - на сервере последняя версия!"
        exit 0
    else
        echo "  ⚠️  Коммиты отличаются:"
        echo "     Локальный:  $LOCAL_HASH"
        echo "     На сервере: $REMOTE_HASH"
        echo ""
        echo "  💡 Для обновления выполните: make deploy"
        exit 1
    fi
elif [ -n "$LOCAL_HASH" ]; then
    echo "  ⚠️  Не удалось получить коммит с сервера"
    echo "     Локальный: $LOCAL_HASH"
    exit 1
else
    echo "  ⚠️  Не git репозиторий локально"
    exit 1
fi

