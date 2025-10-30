package repository

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"lawScraper/scraper/internal/config"
	"lawScraper/scraper/internal/logger"
)

var (
	keywordsMutex sync.RWMutex
	keywordsCache []string
)

// KeywordsData структура для хранения ключевых слов в JSON
type KeywordsData struct {
	Keywords []string `json:"keywords"`
}

// GetKeywordsFilePath возвращает путь к файлу с ключевыми словами
func GetKeywordsFilePath() string {
	return filepath.Join(config.GetProjectRoot(), "data", "keywords.json")
}

// LoadKeywordsFromFile загружает ключевые слова из файла
func LoadKeywordsFromFile() ([]string, error) {
	keywordsMutex.RLock()
	defer keywordsMutex.RUnlock()

	path := GetKeywordsFilePath()
	
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Если файл не существует, возвращаем ключевые слова из переменной окружения
			logger.Log.Info("Файл keywords.json не найден, используем KEYWORDS из .env")
			keywords := config.GetKeywords()
			return keywords, nil
		}
		return nil, err
	}
	
	var keywordsData KeywordsData
	if err := json.Unmarshal(data, &keywordsData); err != nil {
		return nil, err
	}
	
	// Приводим все слова к нижнему регистру
	for i := range keywordsData.Keywords {
		keywordsData.Keywords[i] = strings.ToLower(strings.TrimSpace(keywordsData.Keywords[i]))
	}
	
	return keywordsData.Keywords, nil
}

// SaveKeywordsToFile сохраняет ключевые слова в файл
func SaveKeywordsToFile(keywords []string) error {
	keywordsMutex.Lock()
	defer keywordsMutex.Unlock()

	path := GetKeywordsFilePath()
	
	// Создаем директорию если её нет
	if err := ensureDir(path); err != nil {
		return err
	}
	
	// Приводим все слова к нижнему регистру и удаляем пробелы
	cleanedKeywords := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		cleaned := strings.ToLower(strings.TrimSpace(kw))
		if cleaned != "" {
			cleanedKeywords = append(cleanedKeywords, cleaned)
		}
	}
	
	keywordsData := KeywordsData{
		Keywords: cleanedKeywords,
	}
	
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(keywordsData); err != nil {
		return err
	}
	
	// Обновляем кэш
	keywordsCache = cleanedKeywords
	
	logger.Log.Infof("Ключевые слова сохранены в %s: %v", path, cleanedKeywords)
	return nil
}

// GetCurrentKeywords возвращает текущие ключевые слова (из файла или .env)
func GetCurrentKeywords() []string {
	keywords, err := LoadKeywordsFromFile()
	if err != nil {
		logger.Log.Warnf("Ошибка загрузки ключевых слов из файла: %v, используем .env", err)
		return config.GetKeywords()
	}
	
	if len(keywords) == 0 {
		logger.Log.Info("Файл keywords.json пуст, используем .env")
		return config.GetKeywords()
	}
	
	return keywords
}

// AddKeyword добавляет новое ключевое слово
func AddKeyword(keyword string) error {
	keywords := GetCurrentKeywords()
	
	// Проверяем, нет ли уже такого слова
	cleanedKeyword := strings.ToLower(strings.TrimSpace(keyword))
	for _, kw := range keywords {
		if kw == cleanedKeyword {
			logger.Log.Infof("Ключевое слово '%s' уже существует", cleanedKeyword)
			return nil
		}
	}
	
	keywords = append(keywords, cleanedKeyword)
	return SaveKeywordsToFile(keywords)
}

// RemoveKeyword удаляет ключевое слово
func RemoveKeyword(keyword string) error {
	keywords := GetCurrentKeywords()
	
	cleanedKeyword := strings.ToLower(strings.TrimSpace(keyword))
	newKeywords := make([]string, 0, len(keywords))
	
	for _, kw := range keywords {
		if kw != cleanedKeyword {
			newKeywords = append(newKeywords, kw)
		}
	}
	
	if len(newKeywords) == len(keywords) {
		logger.Log.Infof("Ключевое слово '%s' не найдено", cleanedKeyword)
		return nil
	}
	
	return SaveKeywordsToFile(newKeywords)
}

// SetKeywords заменяет все ключевые слова
func SetKeywords(keywords []string) error {
	return SaveKeywordsToFile(keywords)
}

