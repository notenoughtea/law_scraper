package main

import (
	"fmt"
	"path/filepath"

	"github.com/joho/godotenv"

	"lawScraper/scraper/internal/clients"
	"lawScraper/scraper/internal/config"
	"lawScraper/scraper/internal/logger"
	"lawScraper/scraper/internal/service"
)

func main() {
	logger.Init()
	
	// Загружаем .env из корня проекта
	envPath := filepath.Join(config.GetProjectRoot(), ".env")
	if err := godotenv.Load(envPath); err != nil {
		logger.Log.Warnf("Не удалось загрузить .env: %v", err)
	}

	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Println("║   ТЕСТ ОТПРАВКИ СООБЩЕНИЙ В TELEGRAM              ║")
	fmt.Println("╚════════════════════════════════════════════════════╝")
	fmt.Println()

	// Проверяем конфигурацию
	token := config.GetTelegramToken()
	chatID := config.GetTelegramChatID()

	fmt.Printf("📋 Конфигурация:\n")
	fmt.Printf("   TELEGRAM_BOT_TOKEN: %s\n", maskToken(token))
	fmt.Printf("   TELEGRAM_CHAT_ID: %s\n", chatID)
	fmt.Println()

	if token == "" || chatID == "" {
		fmt.Println("❌ ОШИБКА: Telegram bot token или chat id не настроены!")
		fmt.Println()
		fmt.Println("Пожалуйста, настройте переменные окружения:")
		fmt.Println("  export TELEGRAM_BOT_TOKEN='ваш_токен'")
		fmt.Println("  export TELEGRAM_CHAT_ID='ваш_chat_id'")
		fmt.Println()
		fmt.Println("Или добавьте их в файл .env")
		return
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("1. Тест отправки простого сообщения")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	testMessage := "🧪 <b>Тестовое сообщение</b>\n\nЭто тестовое сообщение для проверки работы Telegram бота."
	
	if err := clients.SendTelegramMessage(testMessage); err != nil {
		fmt.Printf("❌ Ошибка отправки: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("2. Тест отправки уведомлений из файла")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if err := service.SendNotificationsFromFile(); err != nil {
		fmt.Printf("❌ Ошибка: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("✅ ВСЕ ТЕСТЫ ВЫПОЛНЕНЫ УСПЕШНО!")
}

func maskToken(token string) string {
	if token == "" {
		return "<НЕ НАСТРОЕН>"
	}
	if len(token) < 10 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
