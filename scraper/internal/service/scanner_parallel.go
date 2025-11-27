package service

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/notenoughtea/law_scraper/internal/clients"
	"github.com/notenoughtea/law_scraper/internal/config"
	"github.com/notenoughtea/law_scraper/internal/logger"
	"github.com/notenoughtea/law_scraper/internal/repository"
)

var (
	// Максимальное количество одновременных обработчиков файлов
	// По умолчанию 3 для слабого сервера (768MB RAM, 1 CPU)
	// Можно изменить через config.GetMaxWorkers()
	maxWorkers = 3
)

func init() {
	// Загружаем значение из конфига при старте
	maxWorkers = getMaxWorkers()
}

// getMaxWorkers возвращает количество воркеров из конфига
func getMaxWorkers() int {
	if workers := config.GetMaxWorkers(); workers > 0 {
		return workers
	}
	// По умолчанию 3 воркера для слабого сервера
	return 3
}

// fileTask представляет задачу на обработку одного файла
type fileTask struct {
	fileURL     string
	projectURL  string
	projectID   string
	pubDate     string
	title       string
	description string
	keywords    []string
}

// ScanRSSAndProjectsParallel выполняет параллельное сканирование с отправкой уведомлений сразу
// Возвращает количество найденных совпадений
func ScanRSSAndProjectsParallel(rssURL string) (int, error) {
	// Загружаем предыдущий RSS для сравнения
	logger.Log.Info("Загрузка предыдущего RSS для сравнения...")
	oldFeed, err := repository.LoadPreviousRSS()
	if err != nil {
		logger.Log.Warnf("Не удалось загрузить предыдущий RSS: %v", err)
	}

	// Получаем новый RSS
	feed, err := clients.FetchRSS(rssURL)
	if err != nil {
		return 0, err
	}
	logger.Log.Infof("RSS загружен: %d элементов", len(feed.Channel.Items))

	// Получаем только новые элементы
	newItems := repository.GetNewRSSItems(feed, oldFeed)

	if len(newItems) == 0 {
		logger.Log.Info("✓ Новых элементов в RSS не найдено, обработка не требуется")
		return 0, nil
	}

	logger.Log.Infof("🆕 Найдено новых элементов для обработки: %d", len(newItems))

	// Сохраняем новый RSS для следующего запуска
	if err := repository.SaveRSS(feed); err != nil {
		return 0, err
	}
	logger.Log.Info("RSS сохранен для следующего сравнения")

	keywords := repository.LoadKeywords()
	for i := range keywords {
		keywords[i] = strings.ToLower(keywords[i])
	}
	logger.Log.Infof("Ищем ключевые слова: %v", keywords)

	// Канал для задач на обработку файлов
	tasksChan := make(chan fileTask, 100)

	// WaitGroup для синхронизации воркеров
	var wg sync.WaitGroup

	// Счетчик найденных совпадений
	var matchesCount int64
	var matchesMutex sync.Mutex

	// Запускаем воркеры для обработки файлов
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go fileWorker(i+1, tasksChan, keywords, &wg, &matchesCount, &matchesMutex)
	}

	// Собираем все задачи (файлы для обработки)
	totalTasks := 0

	// Обрабатываем каждый новый элемент RSS
	for _, it := range newItems {
		pageURL := it.Link
		html, err := fetch(pageURL)
		if err != nil {
			logger.Log.Warnf("ошибка загрузки страницы %s: %v", pageURL, err)
			continue
		}
		lowerHTML := strings.ToLower(string(html))

		// Проверяем страницу на наличие ключевых слов
		var foundPage []string
		for _, kw := range keywords {
			if kw == "" {
				continue
			}
			if strings.Contains(lowerHTML, kw) {
				foundPage = append(foundPage, kw)
			}
		}

		if len(foundPage) > 0 {
			// Найдено совпадение на странице - отправляем сразу
			logger.Log.Infof("✅ Найдено совпадение на странице %s: %v", pageURL, foundPage)
			sendNotificationImmediately(pageURL, pageURL, foundPage, it.PubDate, it.Title, it.Description, &matchesCount, &matchesMutex)
		}

		// Получаем ID проекта для загрузки файлов
		var projectID string
		if m := projIDRe.FindStringSubmatch(pageURL); len(m) == 2 {
			projectID = m[1]
		}

		if projectID != "" {
			stagesURL := "https://regulation.gov.ru/api/public/PublicProjects/GetProjectStages/" + projectID
			ids, err := clients.FetchProjectStagesFileIDs(stagesURL)
			if err != nil {
				logger.Log.Warnf("ошибка получения стадий проекта %s: %v", projectID, err)
				continue
			}

			// Добавляем задачи на обработку файлов
			for _, fid := range ids {
				fileURL := "https://regulation.gov.ru/api/public/Files/GetFile/" + fid
				tasksChan <- fileTask{
					fileURL:     fileURL,
					projectURL:  pageURL,
					projectID:   projectID,
					pubDate:     it.PubDate,
					title:       it.Title,
					description: it.Description,
				}
				totalTasks++
			}
		}
	}

	// Закрываем канал после отправки всех задач
	close(tasksChan)
	logger.Log.Infof("📋 Всего файлов для обработки: %d", totalTasks)

	// Ждем завершения всех воркеров
	wg.Wait()

	matchesMutex.Lock()
	count := int(matchesCount)
	matchesMutex.Unlock()

	logger.Log.Infof("✅ Все файлы обработаны. Найдено совпадений: %d", count)

	return count, nil
}

// fileWorker обрабатывает файлы из канала задач
func fileWorker(workerID int, tasksChan <-chan fileTask, keywords []string, wg *sync.WaitGroup, matchesCount *int64, matchesMutex *sync.Mutex) {
	defer wg.Done()

	for task := range tasksChan {
		logger.Log.Infof("👷 Воркер %d обрабатывает файл: %s", workerID, task.fileURL)

		// Загружаем файл
		data, err := fetch(task.fileURL)
		if err != nil {
			logger.Log.Warnf("ошибка загрузки вложения %s: %v", task.fileURL, err)
			continue
		}

		// Извлекаем текст из файла
		var textLower string
		if txt, err := extractDocxText(data); err == nil && txt != "" {
			textLower = txt
		} else {
			textLower = decodeToLowerUTF8(data)
		}

		// Ищем ключевые слова
		lower := []byte(textLower)
		var found []string
		for _, kw := range keywords {
			if kw == "" {
				continue
			}
			if bytes.Contains(lower, []byte(kw)) {
				found = append(found, kw)
			}
		}

		// Если найдены совпадения - отправляем уведомление сразу
		if len(found) > 0 {
			logger.Log.Infof("✅ Воркер %d: найдено совпадение в файле %s: %v", workerID, task.fileURL, found)
			sendNotificationImmediately(task.projectURL, task.fileURL, found, task.pubDate, task.title, task.description, matchesCount, matchesMutex)
		} else {
			logger.Log.Debugf("Воркер %d: совпадений не найдено в файле %s", workerID, task.fileURL)
		}

		// Небольшая задержка между файлами для снижения нагрузки
		time.Sleep(100 * time.Millisecond)
	}

	logger.Log.Infof("👷 Воркер %d завершил работу", workerID)
}

// sendNotificationImmediately отправляет уведомление сразу после обработки
func sendNotificationImmediately(projectURL, fileURL string, keywords []string, pubDate, title, description string, matchesCount *int64, matchesMutex *sync.Mutex) {
	// Логируем что передается
	logger.Log.Infof("📤 Отправка уведомления для %s", fileURL)
	logger.Log.Infof("   Ключевые слова: %v (количество: %d)", keywords, len(keywords))
	logger.Log.Infof("   Заголовок: %s", title)

	// Проверка: если keywords пустой, логируем предупреждение
	if len(keywords) == 0 {
		logger.Log.Warnf("⚠️  Ключевые слова пустые для файла %s! Это не должно происходить.", fileURL)
	}

	// Увеличиваем счетчик совпадений
	matchesMutex.Lock()
	*matchesCount++
	count := *matchesCount
	matchesMutex.Unlock()

	// Сохраняем в файл для отслеживания (опционально)
	if fileURL != projectURL {
		// Только для файлов, не для страниц
		fileData := repository.FileURLWithKeywords{
			URL:         fileURL,
			ProjectURL:  projectURL,
			Keywords:    keywords,
			PubDate:     pubDate,
			Title:       title,
			Description: description,
		}

		// Добавляем в файл (аппенд) - с защитой от race condition
		appendToFileURLs(fileData)
	}

	// Отправляем уведомление сразу
	if err := clients.SendFileURLWithKeywords(projectURL, fileURL, keywords, pubDate, title, description); err != nil {
		logger.Log.Errorf("❌ Ошибка отправки уведомления для %s: %v", fileURL, err)
	} else {
		logger.Log.Infof("✅ Уведомление #%d отправлено для %s (ключевые слова: %v)", count, fileURL, keywords)
	}
}

// appendToFileURLs добавляет файл в список (для истории)
// Использует mutex для защиты от race condition
var fileURLsMutex sync.Mutex

func appendToFileURLs(file repository.FileURLWithKeywords) {
	fileURLsMutex.Lock()
	defer fileURLsMutex.Unlock()

	// Загружаем существующий список
	existing, _ := loadFileURLs()

	// Добавляем новый файл
	existing = append(existing, file)

	// Сохраняем обратно
	if err := repository.SaveFileURLs(existing); err != nil {
		logger.Log.Warnf("Не удалось сохранить список URL-ов файлов: %v", err)
	}
}

// loadFileURLs загружает список файлов из JSON
func loadFileURLs() ([]repository.FileURLWithKeywords, error) {
	filePath := filepath.Join(config.GetMatchedDir(), "file_urls.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []repository.FileURLWithKeywords{}, nil
		}
		return nil, err
	}

	var files []repository.FileURLWithKeywords
	if err := json.Unmarshal(data, &files); err != nil {
		return nil, err
	}

	return files, nil
}
