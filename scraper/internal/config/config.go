package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

var projectRoot string

func init() {
	// Определяем корень проекта
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		projectRoot = root
	} else {
		// Получаем абсолютный путь к текущей директории выполнения
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal("Не удалось определить текущую директорию:", err)
		}
		// Ищем корень проекта (где находится go.mod)
		projectRoot = findProjectRoot(cwd)
	}
}

// findProjectRoot ищет корень проекта, поднимаясь вверх по директориям
func findProjectRoot(startPath string) string {
	dir := startPath
	for {
		// Проверяем наличие go.mod или docker-compose.yml
		if fileExists(filepath.Join(dir, "go.mod")) || 
		   fileExists(filepath.Join(dir, "docker-compose.yml")) {
			return dir
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			// Достигли корня файловой системы
			log.Fatal("Не удалось найти корень проекта")
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func GetProjectRoot() string {
	return projectRoot
}

func GetUrl() string {
	value := os.Getenv("API_URL")
	if value == "" {
		log.Fatal("No URL in env")
	}
	return value
}

func GetStoragePath() string {
    if p := os.Getenv("PAGES_STORAGE"); p != "" {
        if filepath.IsAbs(p) {
            return p
        }
        return filepath.Join(projectRoot, p)
    }
    return filepath.Join(projectRoot, "data", "pages.json")
}

func GetKeywords() []string {
    raw := os.Getenv("KEYWORDS")
    if raw == "" {
        return nil
    }
    parts := strings.Split(raw, ",")
    out := make([]string, 0, len(parts))
    for _, p := range parts {
        s := strings.TrimSpace(p)
        if s != "" {
            out = append(out, strings.ToLower(s))
        }
    }
    return out
}

func GetMatchedDir() string {
    if p := os.Getenv("MATCHED_DIR"); p != "" {
        if filepath.IsAbs(p) {
            return p
        }
        return filepath.Join(projectRoot, p)
    }
    return filepath.Join(projectRoot, "data", "matched")
}

func GetTelegramToken() string {
	return os.Getenv("TELEGRAM_BOT_TOKEN")
}

func GetTelegramChatID() string {
	return os.Getenv("TELEGRAM_CHAT_ID")
}

func GetTelegramSendAsDocument() bool {
	// По умолчанию true - отправляем файлы напрямую
	val := os.Getenv("TELEGRAM_SEND_AS_DOCUMENT")
	if val == "" {
		return true // По умолчанию включено
	}
	return val == "true" || val == "1" || val == "yes"
}

func GetCronSchedule() string {
	if s := os.Getenv("CRON_SCHEDULE"); s != "" {
		return s
	}
	return "0 9 * * *" // по умолчанию 9:00 каждый день
}
