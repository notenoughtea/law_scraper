package handler

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"lawScraper/scraper/internal/logger"
	"lawScraper/scraper/internal/repository"
)

type TelegramBotHandler struct {
	bot *tgbotapi.BotAPI
}

// NewTelegramBotHandler создает новый обработчик команд Telegram бота
func NewTelegramBotHandler(bot *tgbotapi.BotAPI) *TelegramBotHandler {
	return &TelegramBotHandler{
		bot: bot,
	}
}

// HandleUpdate обрабатывает входящее обновление от Telegram
func (h *TelegramBotHandler) HandleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message

	// Игнорируем сообщения без текста
	if msg.Text == "" {
		return
	}

	logger.Log.Infof("Получено сообщение от %s: %s", msg.From.UserName, msg.Text)

	// Обрабатываем команды
	if msg.IsCommand() {
		h.handleCommand(msg)
		return
	}

	// Если это не команда, отправляем подсказку
	h.sendHelp(msg.Chat.ID)
}

// handleCommand обрабатывает команды бота
func (h *TelegramBotHandler) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		h.handleStart(msg)
	case "help":
		h.handleHelp(msg)
	case "keywords":
		h.handleKeywords(msg)
	case "set_keywords":
		h.handleSetKeywords(msg)
	case "add_keyword":
		h.handleAddKeyword(msg)
	case "remove_keyword":
		h.handleRemoveKeyword(msg)
	default:
		h.sendMessage(msg.Chat.ID, "❌ Неизвестная команда. Используйте /help для справки.")
	}
}

// handleStart обрабатывает команду /start
func (h *TelegramBotHandler) handleStart(msg *tgbotapi.Message) {
	welcomeText := `👋 <b>Добро пожаловать в Law Scraper Bot!</b>

Я помогу вам управлять ключевыми словами для поиска в нормативных актах.

Используйте /help для просмотра доступных команд.`

	h.sendMessage(msg.Chat.ID, welcomeText)
}

// handleHelp обрабатывает команду /help
func (h *TelegramBotHandler) handleHelp(msg *tgbotapi.Message) {
	h.sendHelp(msg.Chat.ID)
}

// sendHelp отправляет справку по командам
func (h *TelegramBotHandler) sendHelp(chatID int64) {
	helpText := `📚 <b>Доступные команды:</b>

<b>/keywords</b> - показать текущие ключевые слова

<b>/set_keywords</b> слово1,слово2,слово3
   Установить новый список ключевых слов
   Пример: /set_keywords транспорт,образование,здравоохранение

<b>/add_keyword</b> слово
   Добавить новое ключевое слово
   Пример: /add_keyword экология

<b>/remove_keyword</b> слово
   Удалить ключевое слово
   Пример: /remove_keyword транспорт

<b>/help</b> - показать эту справку

<b>📝 Примечания:</b>
• Регистр не важен
• Слова автоматически приводятся к нижнему регистру
• Изменения применяются сразу после команды`

	h.sendMessage(chatID, helpText)
}

// handleKeywords обрабатывает команду /keywords - показать текущие ключевые слова
func (h *TelegramBotHandler) handleKeywords(msg *tgbotapi.Message) {
	keywords := repository.GetCurrentKeywords()

	var response string
	if len(keywords) == 0 {
		response = "❌ Ключевые слова не настроены.\n\nИспользуйте /set_keywords для установки."
	} else {
		keywordsList := strings.Join(keywords, ", ")
		response = fmt.Sprintf("🔑 <b>Текущие ключевые слова (%d):</b>\n\n%s", len(keywords), keywordsList)
	}

	h.sendMessage(msg.Chat.ID, response)
}

// handleSetKeywords обрабатывает команду /set_keywords - установить новый список
func (h *TelegramBotHandler) handleSetKeywords(msg *tgbotapi.Message) {
	args := msg.CommandArguments()
	
	if args == "" {
		h.sendMessage(msg.Chat.ID, "❌ Укажите ключевые слова через запятую.\n\nПример:\n/set_keywords транспорт,образование,здравоохранение")
		return
	}

	// Разбиваем строку на слова
	parts := strings.Split(args, ",")
	keywords := make([]string, 0, len(parts))
	
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			keywords = append(keywords, trimmed)
		}
	}

	if len(keywords) == 0 {
		h.sendMessage(msg.Chat.ID, "❌ Не указано ни одного ключевого слова.")
		return
	}

	// Сохраняем новые ключевые слова
	if err := repository.SetKeywords(keywords); err != nil {
		logger.Log.Errorf("Ошибка сохранения ключевых слов: %v", err)
		h.sendMessage(msg.Chat.ID, fmt.Sprintf("❌ Ошибка сохранения: %v", err))
		return
	}

	keywordsList := strings.Join(keywords, ", ")
	response := fmt.Sprintf("✅ <b>Ключевые слова обновлены (%d):</b>\n\n%s", len(keywords), keywordsList)
	h.sendMessage(msg.Chat.ID, response)
	
	logger.Log.Infof("Пользователь %s установил новые ключевые слова: %v", msg.From.UserName, keywords)
}

// handleAddKeyword обрабатывает команду /add_keyword - добавить одно слово
func (h *TelegramBotHandler) handleAddKeyword(msg *tgbotapi.Message) {
	keyword := strings.TrimSpace(msg.CommandArguments())
	
	if keyword == "" {
		h.sendMessage(msg.Chat.ID, "❌ Укажите ключевое слово.\n\nПример:\n/add_keyword экология")
		return
	}

	// Проверяем, не содержит ли слово запятых (частая ошибка)
	if strings.Contains(keyword, ",") {
		h.sendMessage(msg.Chat.ID, "❌ Команда /add_keyword добавляет только одно слово.\n\nДля нескольких слов используйте:\n/set_keywords слово1,слово2,слово3")
		return
	}

	if err := repository.AddKeyword(keyword); err != nil {
		logger.Log.Errorf("Ошибка добавления ключевого слова: %v", err)
		h.sendMessage(msg.Chat.ID, fmt.Sprintf("❌ Ошибка: %v", err))
		return
	}

	// Показываем обновленный список
	keywords := repository.GetCurrentKeywords()
	keywordsList := strings.Join(keywords, ", ")
	
	response := fmt.Sprintf("✅ <b>Слово '%s' добавлено!</b>\n\n🔑 Текущие ключевые слова (%d):\n%s", 
		strings.ToLower(keyword), len(keywords), keywordsList)
	h.sendMessage(msg.Chat.ID, response)
	
	logger.Log.Infof("Пользователь %s добавил ключевое слово: %s", msg.From.UserName, keyword)
}

// handleRemoveKeyword обрабатывает команду /remove_keyword - удалить слово
func (h *TelegramBotHandler) handleRemoveKeyword(msg *tgbotapi.Message) {
	keyword := strings.TrimSpace(msg.CommandArguments())
	
	if keyword == "" {
		h.sendMessage(msg.Chat.ID, "❌ Укажите ключевое слово для удаления.\n\nПример:\n/remove_keyword транспорт")
		return
	}

	if err := repository.RemoveKeyword(keyword); err != nil {
		logger.Log.Errorf("Ошибка удаления ключевого слова: %v", err)
		h.sendMessage(msg.Chat.ID, fmt.Sprintf("❌ Ошибка: %v", err))
		return
	}

	// Показываем обновленный список
	keywords := repository.GetCurrentKeywords()
	
	var response string
	if len(keywords) == 0 {
		response = fmt.Sprintf("✅ <b>Слово '%s' удалено!</b>\n\n⚠️ Список ключевых слов теперь пуст.", 
			strings.ToLower(keyword))
	} else {
		keywordsList := strings.Join(keywords, ", ")
		response = fmt.Sprintf("✅ <b>Слово '%s' удалено!</b>\n\n🔑 Текущие ключевые слова (%d):\n%s", 
			strings.ToLower(keyword), len(keywords), keywordsList)
	}
	
	h.sendMessage(msg.Chat.ID, response)
	
	logger.Log.Infof("Пользователь %s удалил ключевое слово: %s", msg.From.UserName, keyword)
}

// sendMessage отправляет сообщение в Telegram
func (h *TelegramBotHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	
	if _, err := h.bot.Send(msg); err != nil {
		logger.Log.Errorf("Ошибка отправки сообщения: %v", err)
	}
}

