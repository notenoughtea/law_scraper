package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func SendFileURLWithKeywords(fileURL string, keywords []string, pubDate string) error {
	logger.Log.Infof("📤 Подготовка отправки уведомления для файла: %s", fileURL)
	logger.Log.Infof("Найдено ключевых слов: %d (%v)", len(keywords), keywords)
	logger.Log.Infof("Дата публикации: %s", pubDate)
	
	keywordsStr := ""
	if len(keywords) > 0 {
		keywordsStr = keywords[0]
		for i := 1; i < len(keywords); i++ {
			keywordsStr += ", " + keywords[i]
		}
	}

	message := fmt.Sprintf(
		"🔍 <b>Найдено совпадение</b>\n\n"+
			"📄 <b>Файл:</b> <a href=\"%s\">Ссылка на документ</a>\n"+
			"🔑 <b>Ключевые слова:</b> %s",
		fileURL,
		keywordsStr,
	)
	
	// Добавляем дату публикации если она есть
	if pubDate != "" {
		message += fmt.Sprintf("\n📅 <b>Дата публикации:</b> %s", pubDate)
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

