package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/notenoughtea/law_scraper/internal/clients"
	"github.com/notenoughtea/law_scraper/internal/config"
	"github.com/notenoughtea/law_scraper/internal/logger"
	"github.com/notenoughtea/law_scraper/internal/repository"
)

func SendNotificationsFromFile() error {
	logger.Log.Info("════════════════════════════════════════")
	logger.Log.Info("  НАЧАЛО ПРОЦЕССА ОТПРАВКИ УВЕДОМЛЕНИЙ")
	logger.Log.Info("════════════════════════════════════════")

	filePath := filepath.Join(config.GetMatchedDir(), "file_urls.json")
	logger.Log.Infof("Путь к файлу с URL: %s", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Log.Warn("⚠️  Файл file_urls.json не найден, нечего отправлять")
			return nil
		}
		logger.Log.Errorf("❌ Ошибка чтения файла %s: %v", filePath, err)
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	var files []repository.FileURLWithKeywords
	if err := json.Unmarshal(data, &files); err != nil {
		logger.Log.Errorf("❌ Ошибка парсинга JSON: %v", err)
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	logger.Log.Infof("✓ Загружено %d файлов для отправки", len(files))

	count := 0

	for i, file := range files {
		logger.Log.Infof("────────────────────────────────────────")
		logger.Log.Infof("Обработка файла %d/%d", i+1, len(files))
		logger.Log.Infof("  → URL: %s", file.URL)
		logger.Log.Infof("  → Ключевые слова: %v", file.Keywords)
		logger.Log.Infof("  → Дата публикации: %s", file.PubDate)
		logger.Log.Infof("  → Заголовок: %s", file.Title)
		logger.Log.Infof("  → Описание: %s (длина: %d)",
			truncateString(file.Description, 50), len(file.Description))

		// Отправляем уведомление
		logger.Log.Infof("  → Попытка отправки уведомления %d...", count+1)
		if err := clients.SendFileURLWithKeywords(
			file.ProjectURL,
			file.URL,
			file.Keywords,
			file.PubDate,
			file.Title,
			file.Description,
		); err != nil {
			logger.Log.Errorf("❌ Ошибка отправки уведомления для %s: %v", file.URL, err)
			continue
		}

		count++
		logger.Log.Infof("✅ Уведомление %d отправлено успешно", count)

		// Небольшая задержка между сообщениями, чтобы не превысить лимит Telegram API
		logger.Log.Info("  → Задержка 1 секунда перед следующим сообщением...")
		time.Sleep(1 * time.Second)
	}

	logger.Log.Info("════════════════════════════════════════")
	logger.Log.Infof("  ИТОГО: Отправлено %d уведомлений в Telegram из %d файлов", count, len(files))
	logger.Log.Info("════════════════════════════════════════")
	return nil
}

// truncateString обрезает строку до указанной длины
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// RunManualScan выполняет сканирование вручную и возвращает результат
// Использует параллельную обработку с отправкой уведомлений сразу
func RunManualScan() (int, error) {
	logger.Log.Info("🚀 Запуск ручного сканирования (параллельный режим)...")

	const rssURL = "https://regulation.gov.ru/api/public/Rss/"

	// Используем параллельную версию с отправкой уведомлений сразу
	matchesCount, err := ScanRSSAndProjectsParallel(rssURL)
	if err != nil {
		logger.Log.Errorf("Ошибка сканирования RSS/проектов: %v", err)
		return 0, err
	}

	logger.Log.Infof("✅ Сканирование завершено. Найдено совпадений: %d. Уведомления отправлены сразу.", matchesCount)

	return matchesCount, nil
}
