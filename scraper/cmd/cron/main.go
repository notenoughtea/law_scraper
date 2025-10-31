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

	"github.com/notenoughtea/law_scraper/internal/config"
	"github.com/notenoughtea/law_scraper/internal/handler"
	"github.com/notenoughtea/law_scraper/internal/logger"
	"github.com/notenoughtea/law_scraper/internal/service"
)

func runScanAndNotify() {
	logger.Log.Info("Запуск сканирования (параллельный режим)...")

	const rssURL = "https://regulation.gov.ru/api/public/Rss/"
	
	// Используем параллельную версию с отправкой уведомлений сразу
	// Это оптимизировано для слабых серверов (768MB RAM, 1 CPU)
	matchesCount, err := service.ScanRSSAndProjectsParallel(rssURL)
	if err != nil {
		logger.Log.Errorf("Ошибка сканирования RSS/проектов: %v", err)
		return
	}

	logger.Log.Infof("✅ Задача выполнена успешно. Найдено совпадений: %d. Уведомления отправлены сразу.", matchesCount)
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

