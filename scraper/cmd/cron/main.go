package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"

	"lawScraper/scraper/internal/config"
	"lawScraper/scraper/internal/handler"
	"lawScraper/scraper/internal/logger"
	"lawScraper/scraper/internal/service"
)

func runScanAndNotify() {
	logger.Log.Info("Запуск сканирования...")

	const rssURL = "https://regulation.gov.ru/api/public/Rss/"
	matches, err := service.ScanRSSAndProjects(rssURL)
	if err != nil {
		logger.Log.Errorf("Ошибка сканирования RSS/проектов: %v", err)
		return
	}
	logger.Log.Infof("Найдено совпадений: %d", len(matches))

	// Отправка уведомлений в Telegram
	if err := service.SendNotificationsFromFile(); err != nil {
		logger.Log.Errorf("Ошибка отправки уведомлений: %v", err)
		return
	}

	logger.Log.Info("Задача выполнена успешно")
}

// startTelegramBot запускает Telegram бота для приема команд
func startTelegramBot() {
	token := config.GetTelegramToken()
	if token == "" {
		logger.Log.Warn("TELEGRAM_BOT_TOKEN не установлен, бот не будет запущен")
		return
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Log.Errorf("Ошибка инициализации Telegram бота: %v", err)
		return
	}

	logger.Log.Infof("✅ Telegram бот авторизован: @%s", bot.Self.UserName)

	// Создаем обработчик команд
	botHandler := handler.NewTelegramBotHandler(bot)

	// Настраиваем получение обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	logger.Log.Info("🤖 Telegram бот запущен и ожидает команды...")

	// Обрабатываем входящие обновления
	for update := range updates {
		botHandler.HandleUpdate(update)
	}
}

func main() {
	logger.Init()

	// Загружаем .env из корня проекта
	envPath := filepath.Join(config.GetProjectRoot(), ".env")
	if err := godotenv.Load(envPath); err != nil {
		logger.Log.Warnf("Не удалось загрузить .env: %v (возможно, используются переменные окружения)", err)
	}

	schedule := config.GetCronSchedule()
	logger.Log.Infof("Настройка расписания: %s", schedule)

	c := cron.New()
	
	// Добавляем задачу по расписанию
	_, err := c.AddFunc(schedule, runScanAndNotify)
	if err != nil {
		logger.Log.Fatalf("Ошибка настройки расписания: %v", err)
	}

	// Запуск крон-планировщика
	c.Start()
	logger.Log.Info("Крон-планировщик запущен, ожидание выполнения задач...")

	// Запуск Telegram бота в отдельной горутине
	go startTelegramBot()

	// Опционально: запуск сразу при старте
	if os.Getenv("RUN_ON_START") == "true" {
		logger.Log.Info("RUN_ON_START=true, запуск задачи сразу...")
		runScanAndNotify()
	}

	// Ожидание сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	sig := <-sigChan
	fmt.Printf("\nПолучен сигнал %v, завершение работы...\n", sig)
	
	c.Stop()
	logger.Log.Info("Крон-планировщик остановлен")
}

