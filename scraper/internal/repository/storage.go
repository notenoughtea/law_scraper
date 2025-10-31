package repository

import (
    "encoding/json"
    "errors"
    "io/fs"
    "os"
    "path/filepath"

    "github.com/notenoughtea/law_scraper/internal/config"
    "github.com/notenoughtea/law_scraper/internal/dto"
    "github.com/notenoughtea/law_scraper/internal/logger"
)

func ensureDir(path string) error {
    dir := filepath.Dir(path)
    return os.MkdirAll(dir, 0o755)
}

func SavePages(pages []dto.ListResponse) error {
    storage := config.GetStoragePath()
    if err := ensureDir(storage); err != nil {
        return err
    }
    f, err := os.Create(storage)
    if err != nil {
        return err
    }
    defer f.Close()
    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    return enc.Encode(pages)
}

func LoadPages() ([]dto.ListResponse, error) {
	storage := config.GetStoragePath()
	b, err := os.ReadFile(storage)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var pages []dto.ListResponse
	if err := json.Unmarshal(b, &pages); err != nil {
		return nil, err
	}
	return pages, nil
}

// FileURLWithKeywords содержит URL файла и найденные в нём ключевые слова
type FileURLWithKeywords struct {
	URL         string
	Keywords    []string
	PubDate     string
	Title       string
	Description string
}

// SaveFileURLs сохраняет список URL-ов файлов с найденными словами в JSON файле
func SaveFileURLs(files []FileURLWithKeywords) error {
	dir := config.GetMatchedDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	
	// Сохраняем в JSON формате для удобства
	path := filepath.Join(dir, "file_urls.json")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(files)
}

// ClearPagesData удаляет файл pages.json
func ClearPagesData() error {
	storage := config.GetStoragePath()
	if err := os.Remove(storage); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			logger.Log.Info("Файл pages.json не найден, удалять нечего")
			return nil
		}
		return err
	}
	logger.Log.Info("Файл pages.json успешно удален")
	return nil
}


