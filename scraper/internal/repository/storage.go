package repository

import (
    "encoding/json"
    "errors"
    "io/fs"
    "os"
    "path/filepath"

    "lawScraper/scraper/internal/config"
    "lawScraper/scraper/internal/dto"
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
	URL      string
	Keywords []string
	PubDate  string
}

// SaveFileURLs сохраняет список URL-ов файлов с найденными словами в текстовом файле
// Формат: URL | keywords | pubDate
func SaveFileURLs(files []FileURLWithKeywords) error {
	dir := config.GetMatchedDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, "file_urls.txt")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, file := range files {
		line := file.URL
		
		// Добавляем ключевые слова
		if len(file.Keywords) > 0 {
			line += " | " + file.Keywords[0]
			for i := 1; i < len(file.Keywords); i++ {
				line += ", " + file.Keywords[i]
			}
		} else {
			line += " | "
		}
		
		// Добавляем дату публикации
		if file.PubDate != "" {
			line += " | " + file.PubDate
		}
		
		if _, err := f.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return nil
}


