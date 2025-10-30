package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"

	"lawScraper/scraper/internal/config"
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

