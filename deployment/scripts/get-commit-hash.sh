#!/bin/bash
# Скрипт для получения хеша коммита с сервера

SERVER_HOST="${1:-77.105.133.231}"
SERVER_USER="${2:-root}"

# Пытаемся получить из git репозитория (только если .git существует)
# Подавляем все ошибки, включая "fatal: not a git repository"
GIT_HASH=$(ssh -o StrictHostKeyChecking=no -o ConnectTimeout=10 -o BatchMode=yes "$SERVER_USER@$SERVER_HOST" "cd /opt/law_scraper 2>/dev/null && if [ -d .git ]; then git rev-parse HEAD 2>&1 | grep -v 'fatal:' | head -1; else exit 1; fi" 2>/dev/null || echo "")

if [ -n "$GIT_HASH" ] && echo "$GIT_HASH" | grep -qE '^[a-f0-9]{40}$'; then
    echo "$GIT_HASH"
    exit 0
fi

# Пытаемся получить из .deployment_info
DEPLOY_HASH=$(ssh -o StrictHostKeyChecking=no -o ConnectTimeout=10 -o BatchMode=yes "$SERVER_USER@$SERVER_HOST" "cd /opt/law_scraper 2>/dev/null && if [ -f .deployment_info ]; then grep '^COMMIT_HASH=' .deployment_info | cut -d'=' -f2- | sed 's/^\"//; s/\"\$//' | tr -d '\n\r ' | cut -c1-40; fi" 2>/dev/null | tr -d '\n\r')

if [ -n "$DEPLOY_HASH" ] && echo "$DEPLOY_HASH" | grep -qE '^[a-f0-9]{40}$'; then
    echo "$DEPLOY_HASH"
    exit 0
fi

exit 1

