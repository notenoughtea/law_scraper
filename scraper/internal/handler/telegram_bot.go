package handler

import (
	"fmt"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/notenoughtea/law_scraper/internal/logger"
	"github.com/notenoughtea/law_scraper/internal/repository"
	"github.com/notenoughtea/law_scraper/internal/service"
)

type TelegramBotHandler struct {
	bot          *tgbotapi.BotAPI
	scanMutex    sync.Mutex
	isScanning   bool
}

// NewTelegramBotHandler создает новый обработчик команд Telegram бота
func NewTelegramBotHandler(bot *tgbotapi.BotAPI) *TelegramBotHandler {
	return &TelegramBotHandler{
		bot:        bot,
		isScanning: false,
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
	case "scan":
		h.handleScan(msg)
	case "clear_data":
		h.handleClearData(msg)
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

<b>/scan</b> - запустить парсер вручную
   Начинает сканирование RSS и поиск по ключевым словам

<b>/clear_data</b> - удалить сохраненные данные
   Удаляет rss.json и pages.json (после этого все элементы будут считаться новыми)

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

// handleScan обрабатывает команду /scan - запуск парсера вручную
func (h *TelegramBotHandler) handleScan(msg *tgbotapi.Message) {
	h.scanMutex.Lock()
	if h.isScanning {
		h.scanMutex.Unlock()
		h.sendMessage(msg.Chat.ID, "⏳ Сканирование уже выполняется, пожалуйста подождите...")
		return
	}
	h.isScanning = true
	h.scanMutex.Unlock()

	// Запускаем сканирование в отдельной горутине
	go func() {
		defer func() {
			h.scanMutex.Lock()
			h.isScanning = false
			h.scanMutex.Unlock()
		}()

		h.sendMessage(msg.Chat.ID, "🚀 Запуск сканирования (параллельный режим)...\n\n⏳ Это может занять несколько минут.\n\n💡 <i>Обрабатывается параллельно с ограниченным количеством воркеров для экономии ресурсов.</i>")
		
		matches, err := service.RunManualScan()
		if err != nil {
			h.sendMessage(msg.Chat.ID, fmt.Sprintf("❌ <b>Ошибка сканирования:</b>\n\n%v", err))
			logger.Log.Errorf("Ошибка ручного сканирования: %v", err)
			return
		}

		h.sendMessage(msg.Chat.ID, fmt.Sprintf("✅ <b>Сканирование завершено!</b>\n\n📊 Найдено совпадений: %d\n\n📨 Уведомления отправлены в Telegram сразу после обработки каждого файла.", matches))
		logger.Log.Infof("Пользователь %s запустил ручное сканирование, найдено совпадений: %d", msg.From.UserName, matches)
	}()
}

// handleClearData обрабатывает команду /clear_data - удаление сохраненных данных
func (h *TelegramBotHandler) handleClearData(msg *tgbotapi.Message) {
	// Подтверждение перед удалением
	args := strings.TrimSpace(msg.CommandArguments())
	if args != "yes" {
		h.sendMessage(msg.Chat.ID, "⚠️ <b>Внимание!</b> Эта команда удалит сохраненные данные:\n\n• rss.json - кэш RSS\n• pages.json - кэш страниц\n\nПосле удаления все элементы будут считаться новыми при следующем сканировании.\n\nДля подтверждения отправьте:\n<b>/clear_data yes</b>")
		return
	}

	// Удаляем данные
	rssErr := repository.ClearRSSData()
	pagesErr := repository.ClearPagesData()

	var response string
	if rssErr != nil && pagesErr != nil {
		response = fmt.Sprintf("❌ <b>Ошибки при удалении:</b>\n\nRSS: %v\nPages: %v", rssErr, pagesErr)
	} else if rssErr != nil {
		response = fmt.Sprintf("⚠️ <b>Частично удалено:</b>\n\n✅ pages.json удален\n❌ rss.json: %v", rssErr)
	} else if pagesErr != nil {
		response = fmt.Sprintf("⚠️ <b>Частично удалено:</b>\n\n✅ rss.json удален\n❌ pages.json: %v", pagesErr)
	} else {
		response = "✅ <b>Данные успешно удалены!</b>\n\n• rss.json\n• pages.json\n\nПри следующем сканировании все элементы будут считаться новыми."
	}

	h.sendMessage(msg.Chat.ID, response)
	logger.Log.Infof("Пользователь %s очистил сохраненные данные", msg.From.UserName)
}

// sendMessage отправляет сообщение в Telegram
func (h *TelegramBotHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	
	if _, err := h.bot.Send(msg); err != nil {
		logger.Log.Errorf("Ошибка отправки сообщения: %v", err)
	}
}

