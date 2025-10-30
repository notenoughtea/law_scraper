package service

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"lawScraper/scraper/internal/clients"
	"lawScraper/scraper/internal/config"
	"lawScraper/scraper/internal/logger"
)

func SendNotificationsFromFile() error {
	logger.Log.Info("════════════════════════════════════════")
	logger.Log.Info("  НАЧАЛО ПРОЦЕССА ОТПРАВКИ УВЕДОМЛЕНИЙ")
	logger.Log.Info("════════════════════════════════════════")
	
	filePath := config.GetMatchedDir() + "/file_urls.txt"
	logger.Log.Infof("Путь к файлу с URL: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Log.Warn("⚠️  Файл file_urls.txt не найден, нечего отправлять")
			return nil
		}
		logger.Log.Errorf("❌ Ошибка открытия файла %s: %v", filePath, err)
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	logger.Log.Info("✓ Файл успешно открыт, начинаем чтение...")

	scanner := bufio.NewScanner(file)
	count := 0
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		logger.Log.Infof("────────────────────────────────────────")
		logger.Log.Infof("Обработка строки %d: %s", lineNum, line)
		
		if line == "" {
			logger.Log.Infof("  → Пустая строка, пропускаем")
			continue
		}

		// Парсим строку: URL | keyword1, keyword2, ... | pubDate
		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			logger.Log.Warnf("⚠️  Некорректная строка в файле (нет разделителя '|'): %s", line)
			continue
		}

		fileURL := strings.TrimSpace(parts[0])
		keywordsStr := strings.TrimSpace(parts[1])
		pubDate := ""
		if len(parts) >= 3 {
			pubDate = strings.TrimSpace(parts[2])
		}
		
		keywords := make([]string, 0)
		if keywordsStr != "" {
			for _, kw := range strings.Split(keywordsStr, ",") {
				keywords = append(keywords, strings.TrimSpace(kw))
			}
		}

		logger.Log.Infof("  → URL: %s", fileURL)
		logger.Log.Infof("  → Ключевые слова: %v", keywords)
		logger.Log.Infof("  → Дата публикации: %s", pubDate)

		// Отправляем уведомление
		logger.Log.Infof("  → Попытка отправки уведомления %d...", count+1)
		if err := clients.SendFileURLWithKeywords(fileURL, keywords, pubDate); err != nil {
			logger.Log.Errorf("❌ Ошибка отправки уведомления для %s: %v", fileURL, err)
			continue
		}

		count++
		logger.Log.Infof("✅ Уведомление %d отправлено успешно", count)
		
		// Небольшая задержка между сообщениями, чтобы не превысить лимит Telegram API
		logger.Log.Info("  → Задержка 1 секунда перед следующим сообщением...")
		time.Sleep(1 * time.Second)
	}

	if err := scanner.Err(); err != nil {
		logger.Log.Errorf("❌ Ошибка чтения файла: %v", err)
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	logger.Log.Info("════════════════════════════════════════")
	logger.Log.Infof("  ИТОГО: Отправлено %d уведомлений в Telegram из %d строк", count, lineNum)
	logger.Log.Info("════════════════════════════════════════")
	return nil
}

