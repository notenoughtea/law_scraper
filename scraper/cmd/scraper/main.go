package main

import (
	"path/filepath"

	"lawScraper/scraper/internal/clients"
	"lawScraper/scraper/internal/config"
	"lawScraper/scraper/internal/logger"
	"lawScraper/scraper/internal/repository"
	"lawScraper/scraper/internal/service"

	"github.com/joho/godotenv"
)

func main() {
	logger.Init()

	// Загружаем .env из корня проекта
	envPath := filepath.Join(config.GetProjectRoot(), ".env")
	err := godotenv.Load(envPath)
	if err != nil {
		logger.Log.Panicf("ошибка загрузки .env: %v", err)
	}
	// 1) Поддержка старого кэша (при наличии)
	if cached, err := repository.LoadPages(); err == nil && len(cached) > 0 {
		logger.Log.Infof("Найден кэш страниц (legacy): %d", len(cached))
	}

	// 2) Новый поток: RSS -> проекты -> вложения -> KEYWORDS
	const rssURL = "https://regulation.gov.ru/api/public/Rss/"
	matches, err := service.ScanRSSAndProjects(rssURL)
	if err != nil {
		logger.Log.Panicf("ошибка сканирования RSS/проектов: %v", err)
	}
	logger.Log.Infof("Найдено совпадений: %d", len(matches))
	for _, m := range matches {
		logger.Log.Infof("Проект: %s, файл: %s, ключи: %v", m.ProjectURL, m.FileURL, m.Keywords)
	}

	// Сохранение первых 5 старых страниц для совместимости
	if _, err := clients.GetActsList(); err != nil {
		logger.Log.Warnf("legacy загрузка страниц не выполнена: %v", err)
	}
}
