package clients

import (
    "encoding/xml"
    "io"
    "net/http"

    "github.com/notenoughtea/law_scraper/internal/dto"
)

func FetchRSS(url string) (*dto.RSS, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Accept", "application/rss+xml, application/xml, */*")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    b, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    var feed dto.RSS
    if err := xml.Unmarshal(b, &feed); err != nil {
        return nil, err
    }
    return &feed, nil
}


