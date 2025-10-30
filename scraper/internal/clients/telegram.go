package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"lawScraper/scraper/internal/config"
	"lawScraper/scraper/internal/logger"
)

type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func SendTelegramMessage(message string) error {
	logger.Log.Info("=== Начало отправки сообщения в Telegram ===")
	
	token := config.GetTelegramToken()
	chatID := config.GetTelegramChatID()

	logger.Log.Infof("Проверка конфигурации: Token=%s, ChatID=%s", 
		maskToken(token), chatID)

	if token == "" || chatID == "" {
		logger.Log.Error("❌ Telegram bot token или chat id не настроены!")
		logger.Log.Errorf("Token пустой: %v, ChatID пустой: %v", token == "", chatID == "")
		return fmt.Errorf("telegram bot token или chat id не настроены")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	logger.Log.Infof("URL для отправки: https://api.telegram.org/bot***HIDDEN***/sendMessage")

	msg := TelegramMessage{
		ChatID:    chatID,
		Text:      message,
		ParseMode: "HTML",
	}

	logger.Log.Infof("Формирование сообщения: ChatID=%s, Длина текста=%d, ParseMode=%s", 
		chatID, len(message), "HTML")
	logger.Log.Debugf("Текст сообщения: %s", message)

	jsonData, err := json.Marshal(msg)
	if err != nil {
		logger.Log.Errorf("❌ Ошибка сериализации сообщения: %v", err)
		return fmt.Errorf("ошибка сериализации сообщения: %w", err)
	}
	logger.Log.Infof("✓ Сообщение сериализовано, размер JSON: %d байт", len(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Log.Errorf("❌ Ошибка создания HTTP запроса: %v", err)
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	logger.Log.Info("✓ HTTP запрос создан")

	logger.Log.Info("Отправка запроса в Telegram API...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Errorf("❌ Ошибка выполнения HTTP запроса: %v", err)
		return fmt.Errorf("ошибка отправки сообщения: %w", err)
	}
	defer resp.Body.Close()

	logger.Log.Infof("Получен ответ от Telegram API: Status=%d (%s)", 
		resp.StatusCode, resp.Status)

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		logger.Log.Warnf("Не удалось прочитать тело ответа: %v", readErr)
	} else {
		logger.Log.Infof("Тело ответа от Telegram API: %s", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		logger.Log.Errorf("❌ Telegram API вернул ошибку: %s", resp.Status)
		return fmt.Errorf("telegram api вернул ошибку: %s, тело ответа: %s", resp.Status, string(body))
	}

	logger.Log.Info("✅ Сообщение успешно отправлено в Telegram")
	logger.Log.Info("=== Конец отправки сообщения в Telegram ===")
	return nil
}

// maskToken маскирует токен для безопасного логирования
func maskToken(token string) string {
	if token == "" {
		return "<ПУСТО>"
	}
	if len(token) < 10 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func SendFileURLWithKeywords(fileURL string, keywords []string, pubDate string, title string, description string) error {
	logger.Log.Infof("📤 Подготовка отправки уведомления для файла: %s", fileURL)
	logger.Log.Infof("Найдено ключевых слов: %d (%v)", len(keywords), keywords)
	logger.Log.Infof("Дата публикации: %s", pubDate)
	logger.Log.Infof("Заголовок: %s", title)
	
	keywordsStr := ""
	if len(keywords) > 0 {
		keywordsStr = keywords[0]
		for i := 1; i < len(keywords); i++ {
			keywordsStr += ", " + keywords[i]
		}
	}

	// Формируем caption для документа
	caption := "🔍 <b>Найдено совпадение</b>\n\n"
	
	if title != "" {
		caption += fmt.Sprintf("📋 <b>%s</b>\n\n", title)
	}
	
	if description != "" {
		// Ограничиваем длину description для Telegram (макс 1024 символа для caption)
		maxDescLen := 500
		desc := description
		if len(desc) > maxDescLen {
			desc = desc[:maxDescLen] + "..."
		}
		caption += fmt.Sprintf("📝 %s\n\n", desc)
	}
	
	caption += fmt.Sprintf("🔑 <b>Ключевые слова:</b> %s", keywordsStr)
	
	if pubDate != "" {
		caption += fmt.Sprintf("\n📅 <b>Дата:</b> %s", pubDate)
	}

	// Проверяем режим отправки (отправлять ли файл напрямую)
	sendAsDocument := config.GetTelegramSendAsDocument()
	
	if sendAsDocument {
		logger.Log.Info("Режим: отправка файла как документ в Telegram")
		// Отправляем файл напрямую как документ
		return SendDocumentToTelegram(fileURL, caption)
	}

	// Режим по умолчанию: отправка ссылки на файл
	logger.Log.Info("Режим: отправка ссылки на файл")
	
	message := "🔍 <b>Найдено совпадение</b>\n\n"
	
	if title != "" {
		message += fmt.Sprintf("📋 <b>%s</b>\n\n", title)
	}
	
	message += fmt.Sprintf("📄 <b>Файл:</b> <a href=\"%s\">Скачать документ</a>\n", fileURL)
	
	if description != "" {
		// Ограничиваем длину description
		maxDescLen := 500
		desc := description
		if len(desc) > maxDescLen {
			desc = desc[:maxDescLen] + "..."
		}
		message += fmt.Sprintf("📝 %s\n\n", desc)
	}
	
	message += fmt.Sprintf("🔑 <b>Ключевые слова:</b> %s", keywordsStr)
	
	if pubDate != "" {
		message += fmt.Sprintf("\n📅 <b>Дата:</b> %s", pubDate)
	}
	
	// Добавляем инструкцию о расширении файла
	if !hasExtension(fileURL) {
		message += "\n\n💡 <i>После скачивания переименуйте файл, добавив расширение .docx</i>"
	}

	logger.Log.Infof("Сформированное сообщение для отправки (длина: %d символов)", len(message))
	
	err := SendTelegramMessage(message)
	if err != nil {
		logger.Log.Errorf("❌ Ошибка отправки уведомления для %s: %v", fileURL, err)
		return err
	}
	
	logger.Log.Infof("✅ Уведомление для %s отправлено успешно", fileURL)
	return nil
}

// hasExtension проверяет, есть ли расширение файла в URL
func hasExtension(url string) bool {
	// Проверяем наличие типичных расширений документов
	extensions := []string{".docx", ".doc", ".pdf", ".txt", ".xlsx", ".xls"}
	for _, ext := range extensions {
		if len(url) >= len(ext) && url[len(url)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// SendDocumentToTelegram отправляет файл как документ в Telegram
func SendDocumentToTelegram(fileURL string, caption string) error {
	token := config.GetTelegramToken()
	chatID := config.GetTelegramChatID()

	if token == "" || chatID == "" {
		return fmt.Errorf("telegram bot token или chat id не настроены")
	}

	logger.Log.Infof("Скачивание файла с %s...", fileURL)
	
	// Скачиваем файл
	resp, err := http.Get(fileURL)
	if err != nil {
		return fmt.Errorf("ошибка скачивания файла: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка при скачивании файла: статус %d", resp.StatusCode)
	}

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	logger.Log.Infof("Файл скачан, размер: %d байт", len(fileData))

	// Отправляем файл в Telegram
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", token)

	// Создаем multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Добавляем chat_id
	_ = writer.WriteField("chat_id", chatID)
	
	// Добавляем caption
	if caption != "" {
		_ = writer.WriteField("caption", caption)
		_ = writer.WriteField("parse_mode", "HTML")
	}

	// Добавляем файл
	part, err := writer.CreateFormFile("document", "document.docx")
	if err != nil {
		return fmt.Errorf("ошибка создания form file: %w", err)
	}
	
	if _, err := part.Write(fileData); err != nil {
		return fmt.Errorf("ошибка записи файла: %w", err)
	}

	writer.Close()

	// Отправляем запрос
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	logger.Log.Info("Отправка документа в Telegram...")
	
	apiResp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка отправки документа: %w", err)
	}
	defer apiResp.Body.Close()

	respBody, _ := io.ReadAll(apiResp.Body)
	
	if apiResp.StatusCode != http.StatusOK {
		logger.Log.Errorf("❌ Ошибка Telegram API: %s, тело: %s", apiResp.Status, string(respBody))
		return fmt.Errorf("telegram api вернул ошибку: %s", apiResp.Status)
	}

	logger.Log.Info("✅ Документ успешно отправлен в Telegram")
	return nil
}

