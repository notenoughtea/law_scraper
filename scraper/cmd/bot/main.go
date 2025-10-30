package main

import (
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"lawScraper/scraper/internal/config"
	"lawScraper/scraper/internal/handler"
	"lawScraper/scraper/internal/logger"
)

func main() {
	logger.Init()
	logger.Log.Info("════════════════════════════════════════")
	logger.Log.Info("  ЗАПУСК TELEGRAM БОТА ДЛЯ УПРАВЛЕНИЯ")
	logger.Log.Info("════════════════════════════════════════")

	// Загружаем .env из корня проекта
	envPath := filepath.Join(config.GetProjectRoot(), ".env")
	if err := godotenv.Load(envPath); err != nil {
		logger.Log.Warnf("Не удалось загрузить .env: %v (возможно, используются переменные окружения)", err)
	}

	// Получаем токен бота
	token := config.GetTelegramToken()
	if token == "" {
		logger.Log.Fatal("TELEGRAM_BOT_TOKEN не установлен в переменных окружения")
	}

	// Создаем бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Log.Fatalf("Ошибка создания бота: %v", err)
	}

	logger.Log.Infof("✅ Бот успешно авторизован: @%s", bot.Self.UserName)

	// Включаем режим отладки (опционально)
	debug := os.Getenv("TELEGRAM_DEBUG") == "true"
	bot.Debug = debug
	if debug {
		logger.Log.Info("🔧 Режим отладки Telegram Bot API включен")
	}

	// Создаем обработчик команд
	botHandler := handler.NewTelegramBotHandler(bot)

	// Настраиваем получение обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	logger.Log.Info("🚀 Бот запущен и ожидает команды...")
	logger.Log.Info("════════════════════════════════════════")
	logger.Log.Info("")
	logger.Log.Info("📋 Доступные команды:")
	logger.Log.Info("  /start - приветствие")
	logger.Log.Info("  /help - справка")
	logger.Log.Info("  /keywords - показать текущие ключевые слова")
	logger.Log.Info("  /set_keywords слово1,слово2 - установить новые слова")
	logger.Log.Info("  /add_keyword слово - добавить слово")
	logger.Log.Info("  /remove_keyword слово - удалить слово")
	logger.Log.Info("")
	logger.Log.Info("════════════════════════════════════════")

	// Обрабатываем входящие обновления
	for update := range updates {
		// Обрабатываем каждое обновление в отдельной горутине
		go botHandler.HandleUpdate(update)
	}
}

