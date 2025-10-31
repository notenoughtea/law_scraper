package service

import (
    "bytes"
    "errors"
    "io"
    "net/http"
    "net/url"
    "path"
    "regexp"
    "strings"
    "unicode/utf8"
    "archive/zip"
    "encoding/xml"
    "fmt"

    "github.com/notenoughtea/law_scraper/internal/clients"
    "github.com/notenoughtea/law_scraper/internal/logger"
    "github.com/notenoughtea/law_scraper/internal/repository"
)

var attachmentRe = regexp.MustCompile(`href="([^"]+\.(?:pdf|txt|docx|doc))"`)
var projIDRe = regexp.MustCompile(`/projects/(\d+)`)

func absoluteLink(base string, href string) string {
    u, err := url.Parse(href)
    if err != nil {
        return href
    }
    if u.IsAbs() {
        return href
    }
    bu, err := url.Parse(base)
    if err != nil {
        return href
    }
    // join relative
    bu.Path = path.Join(bu.Path, u.Path)
    return bu.String()
}

func fetch(url string) ([]byte, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Accept", "*/*")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, errors.New(resp.Status)
    }
    return io.ReadAll(resp.Body)
}

func isMostlyPrintable(s string) bool {
    if len(s) == 0 { return false }
    total := 0
    printable := 0
    for _, r := range s {
        total++
        if r == '\n' || r == '\r' || r == '\t' || (r >= 32 && r < 127) || (r >= 0x0400 && r <= 0x04FF) {
            printable++
        }
    }
    return printable*100/total >= 80
}

// extractDocxText извлекает текст из DOCX (word/document.xml) и возвращает его в нижнем регистре UTF-8
func extractDocxText(data []byte) (string, error) {
    zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
    if err != nil {
        return "", err
    }
    var r io.ReadCloser
    for _, f := range zr.File {
        if f.Name == "word/document.xml" {
            rc, err := f.Open()
            if err != nil {
                return "", err
            }
            r = rc
            break
        }
    }
    if r == nil {
        return "", fmt.Errorf("document.xml not found")
    }
    defer r.Close()
    dec := xml.NewDecoder(r)
    var b strings.Builder
    inText := false
    for {
        tok, err := dec.Token()
        if err == io.EOF {
            break
        }
        if err != nil {
            return "", err
        }
        switch t := tok.(type) {
        case xml.StartElement:
            if t.Name.Local == "t" { // w:t
                inText = true
            }
        case xml.EndElement:
            if t.Name.Local == "t" {
                inText = false
                b.WriteByte(' ')
            }
        case xml.CharData:
            if inText {
                b.WriteString(string(t))
            }
        }
    }
    return strings.ToLower(b.String()), nil
}

// decodeWindows1251 декодирует байты cp1251 в строку UTF-8
func decodeWindows1251(b []byte) string {
    // Таблица соответствий 0x80-0xFF → Unicode (cp1251)
    var cp1251 = [128]rune{
        0x0402, 0x0403, 0x201A, 0x0453, 0x201E, 0x2026, 0x2020, 0x2021,
        0x20AC, 0x2030, 0x0409, 0x2039, 0x040A, 0x040C, 0x040B, 0x040F,
        0x0452, 0x2018, 0x2019, 0x201C, 0x201D, 0x2022, 0x2013, 0x2014,
        0x00,   0x2122, 0x0459, 0x203A, 0x045A, 0x045C, 0x045B, 0x045F,
        0x00A0, 0x040E, 0x045E, 0x0408, 0x00A4, 0x0490, 0x00A6, 0x00A7,
        0x0401, 0x00A9, 0x0404, 0x00AB, 0x00AC, 0x00AD, 0x00AE, 0x0407,
        0x00B0, 0x00B1, 0x0406, 0x0456, 0x0491, 0x00B5, 0x00B6, 0x00B7,
        0x0451, 0x2116, 0x0454, 0x00BB, 0x0458, 0x0405, 0x0455, 0x0457,
        0x0410, 0x0411, 0x0412, 0x0413, 0x0414, 0x0415, 0x0416, 0x0417,
        0x0418, 0x0419, 0x041A, 0x041B, 0x041C, 0x041D, 0x041E, 0x041F,
        0x0420, 0x0421, 0x0422, 0x0423, 0x0424, 0x0425, 0x0426, 0x0427,
        0x0428, 0x0429, 0x042A, 0x042B, 0x042C, 0x042D, 0x042E, 0x042F,
        0x0430, 0x0431, 0x0432, 0x0433, 0x0434, 0x0435, 0x0436, 0x0437,
        0x0438, 0x0439, 0x043A, 0x043B, 0x043C, 0x043D, 0x043E, 0x043F,
        0x0440, 0x0441, 0x0442, 0x0443, 0x0444, 0x0445, 0x0446, 0x0447,
        0x0448, 0x0449, 0x044A, 0x044B, 0x044C, 0x044D, 0x044E, 0x044F,
    }
    out := make([]rune, 0, len(b))
    for _, by := range b {
        if by < 0x80 {
            out = append(out, rune(by))
        } else {
            r := cp1251[by-0x80]
            if r == 0x00 {
                // неизвестный символ – пропустим
                continue
            }
            out = append(out, r)
        }
    }
    return strings.ToLower(string(out))
}

// decodeToLowerUTF8 пытается вернуть текст в UTF-8 (нижний регистр):
// - если вход валиден как UTF-8 — используем его;
// - иначе пробуем cp1251 → UTF-8.
func decodeToLowerUTF8(b []byte) string {
    if utf8.Valid(b) {
        return strings.ToLower(string(b))
    }
    return decodeWindows1251(b)
}

type Match struct {
    ProjectURL  string   `json:"projectUrl"`
    FileURL     string   `json:"fileUrl"`
    Keywords    []string `json:"keywords"`
    PubDate     string   `json:"pubDate"`     // Дата публикации из RSS
    Title       string   `json:"title"`       // Заголовок из RSS
    Description string   `json:"description"` // Описание из RSS
}

func ScanRSSAndProjects(rssURL string) ([]Match, error) {
    // Загружаем предыдущий RSS для сравнения
    logger.Log.Info("Загрузка предыдущего RSS для сравнения...")
    oldFeed, err := repository.LoadPreviousRSS()
    if err != nil {
        logger.Log.Warnf("Не удалось загрузить предыдущий RSS: %v", err)
    }
    
    // Получаем новый RSS
    feed, err := clients.FetchRSS(rssURL)
    if err != nil {
        return nil, err
    }
    logger.Log.Infof("RSS загружен: %d элементов", len(feed.Channel.Items))
    
    // Получаем только новые элементы
    newItems := repository.GetNewRSSItems(feed, oldFeed)
    
    if len(newItems) == 0 {
        logger.Log.Info("✓ Новых элементов в RSS не найдено, обработка не требуется")
        return []Match{}, nil
    }
    
    logger.Log.Infof("🆕 Найдено новых элементов для обработки: %d", len(newItems))
    
    // Сохраняем новый RSS для следующего запуска
    if err := repository.SaveRSS(feed); err != nil {
        return nil, err
    }
    logger.Log.Info("RSS сохранен для следующего сравнения")

    keywords := repository.LoadKeywords()
    for i := range keywords {
        keywords[i] = strings.ToLower(keywords[i])
    }
    logger.Log.Infof("Ищем ключевые слова: %v", keywords)

    var matches []Match
    for _, it := range newItems {
        pageURL := it.Link
        html, err := fetch(pageURL)
        if err != nil {
            logger.Log.Warnf("ошибка загрузки страницы %s: %v", pageURL, err)
            continue
        }
        lowerHTML := strings.ToLower(string(html))

        // 1) искать совпадения прямо в HTML страницы
        var foundPage []string
        for _, kw := range keywords {
            if kw == "" { continue }
            if strings.Contains(lowerHTML, kw) {
                foundPage = append(foundPage, kw)
            }
        }
        if len(foundPage) > 0 {
            matches = append(matches, Match{
                ProjectURL:  pageURL, 
                FileURL:     pageURL, 
                Keywords:    foundPage,
                PubDate:     it.PubDate,
                Title:       it.Title,
                Description: it.Description,
            })
        }

        // 2) предпочтительный способ: получить ID файлов через GetProjectStages/{id}
        var projectID string
        if m := projIDRe.FindStringSubmatch(pageURL); len(m) == 2 {
            projectID = m[1]
        }
        if projectID != "" {
            stagesURL := "https://regulation.gov.ru/api/public/PublicProjects/GetProjectStages/" + projectID
            ids, err := clients.FetchProjectStagesFileIDs(stagesURL)
            if err != nil {
                logger.Log.Warnf("ошибка получения стадий проекта %s: %v", projectID, err)
            } else {
                for _, fid := range ids {
                    fileURL := "https://regulation.gov.ru/api/public/Files/GetFile/" + fid
                    data, err := fetch(fileURL)
                    if err != nil {
                        logger.Log.Warnf("ошибка загрузки вложения %s: %v", fileURL, err)
                        continue
                    }
                    // Попытка распарсить как DOCX; если не получилось — декодировать как текст
                    var textLower string
                    if txt, err := extractDocxText(data); err == nil && txt != "" {
                        textLower = txt
                    } else {
                        textLower = decodeToLowerUTF8(data)
                    }
                    // Лог содержимого только если оно похоже на читаемый текст
                    if isMostlyPrintable(textLower) {
                        logger.Log.Infof("содержимое файла (utf-8):\n%s", textLower)
                    } else {
                        logger.Log.Infof("содержимое файла: бинарное или нечитаемое, текст опущен")
                    }
                    lower := []byte(textLower)
                    var found []string
                    for _, kw := range keywords {
                        if kw == "" { continue }
                        if bytes.Contains(lower, []byte(kw)) {
                            logger.Log.Infof("сравнение слова: файл=%s, слово='%s' -> найдено", fileURL, kw)
                            found = append(found, kw)
                        } else {
                            logger.Log.Infof("сравнение слова: файл=%s, слово='%s' -> нет", fileURL, kw)
                        }
                    }
                    if len(found) > 0 {
                        logger.Log.Infof("сравнение слов: файл=%s, найдено=%v", fileURL, found)
                        matches = append(matches, Match{
                            ProjectURL:  pageURL, 
                            FileURL:     fileURL, 
                            Keywords:    found,
                            PubDate:     it.PubDate,
                            Title:       it.Title,
                            Description: it.Description,
                        })
                    } else {
                        logger.Log.Infof("сравнение слов: файл=%s, совпадений нет", fileURL)
                    }
                }
            }
		}
	}
	
	// Собрать и сохранить список URL-ов файлов с ключевыми словами
	fileURLs := make([]repository.FileURLWithKeywords, 0)
	for _, m := range matches {
		if m.FileURL != m.ProjectURL {
			// Добавляем только URL-ы файлов, а не страниц проектов
			fileURLs = append(fileURLs, repository.FileURLWithKeywords{
				URL:         m.FileURL,
				Keywords:    m.Keywords,
				PubDate:     m.PubDate,
				Title:       m.Title,
				Description: m.Description,
			})
		}
	}
	if len(fileURLs) > 0 {
		if err := repository.SaveFileURLs(fileURLs); err != nil {
			logger.Log.Warnf("не удалось сохранить список URL-ов файлов: %v", err)
		} else {
			logger.Log.Infof("сохранен список из %d URL-ов файлов с метаданными", len(fileURLs))
		}
	}
	
	return matches, nil
}


