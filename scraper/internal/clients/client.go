package clients

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/notenoughtea/law_scraper/internal/config"
	"github.com/notenoughtea/law_scraper/internal/dto"
	"github.com/notenoughtea/law_scraper/internal/logger"
	"github.com/notenoughtea/law_scraper/internal/repository"
)

func GetActsList() ([]dto.ListResponse, error) {
	url := config.GetUrl()

	type filterModel struct {
		Filters  string `json:"filters"`
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
	}
	type listParams struct {
		FilterModel filterModel `json:"filterModel"`
	}
	type requestPayload struct {
		ListParams    listParams `json:"listParams"`
		OrderedFields []string   `json:"orderedFields"`
	}

	ordered := []string{
		"id",
		"npaStatistics",
		"title",
		"startPublicDiscussion",
		"endPublicDiscussion",
		"okveds",
		"developedDepartment",
		"stage",
		"status",
		"procedure",
	}

	client := &http.Client{}
	var pages []dto.ListResponse

	for p := 1; p <= 5; p++ {
		payload := requestPayload{
			ListParams: listParams{
				FilterModel: filterModel{
					Filters:  "",
					Page:     p,
					PageSize: 20,
				},
			},
			OrderedFields: ordered,
		}

		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Mobile Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		var pageResp dto.ListResponse
		if err := json.Unmarshal(b, &pageResp); err != nil {
			return nil, err
		}
		pages = append(pages, pageResp)
	}

	if err := repository.SavePages(pages); err != nil {
		return nil, err
	}
	logger.Log.Infof("Страницы сохранены: %d", len(pages))
	return pages, nil
}
