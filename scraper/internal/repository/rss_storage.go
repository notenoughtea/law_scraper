package repository

import (
    "encoding/json"
    "errors"
    "io/fs"
    "os"
    "path/filepath"

    "lawScraper/scraper/internal/config"
    "lawScraper/scraper/internal/dto"
    "lawScraper/scraper/internal/logger"
)

func SaveRSS(feed *dto.RSS) error {
    path := filepath.Join(config.GetProjectRoot(), "data", "rss.json")
    if err := ensureDir(path); err != nil {
        return err
    }
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()
    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    return enc.Encode(feed)
}

// LoadPreviousRSS загружает предыдущий сохранённый RSS
func LoadPreviousRSS() (*dto.RSS, error) {
    path := filepath.Join(config.GetProjectRoot(), "data", "rss.json")
    
    data, err := os.ReadFile(path)
    if err != nil {
        if errors.Is(err, fs.ErrNotExist) {
            logger.Log.Info("Предыдущий RSS файл не найден, это первый запуск")
            return nil, nil
        }
        return nil, err
    }
    
    var feed dto.RSS
    if err := json.Unmarshal(data, &feed); err != nil {
        return nil, err
    }
    
    return &feed, nil
}

// GetNewRSSItems возвращает только новые элементы RSS, которых нет в старом
func GetNewRSSItems(newFeed, oldFeed *dto.RSS) []dto.RSSItem {
    if oldFeed == nil || len(oldFeed.Channel.Items) == 0 {
        logger.Log.Info("Нет предыдущего RSS, все элементы считаются новыми")
        return newFeed.Channel.Items
    }
    
    // Создаём мапу для быстрого поиска старых ссылок
    oldLinks := make(map[string]bool)
    for _, item := range oldFeed.Channel.Items {
        oldLinks[item.Link] = true
    }
    
    // Фильтруем только новые элементы
    var newItems []dto.RSSItem
    for _, item := range newFeed.Channel.Items {
        if !oldLinks[item.Link] {
            newItems = append(newItems, item)
        }
    }
    
    logger.Log.Infof("Всего элементов в RSS: %d, новых: %d, старых (пропущено): %d", 
        len(newFeed.Channel.Items), len(newItems), len(oldFeed.Channel.Items))
    
    return newItems
}

func LoadKeywords() []string {
    kws := config.GetKeywords()
    return kws
}


