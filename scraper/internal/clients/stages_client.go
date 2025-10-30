package clients

import (
    "encoding/json"
    "io"
    "net/http"
    "regexp"
)

var uuidRe = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// FetchProjectStagesFileIDs получает JSON по GetProjectStages/{id} и извлекает все UUID-подобные идентификаторы файлов
func FetchProjectStagesFileIDs(apiURL string) ([]string, error) {
    req, err := http.NewRequest("GET", apiURL, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Accept", "application/json, text/plain, */*")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    b, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var any interface{}
    if err := json.Unmarshal(b, &any); err != nil {
        return nil, err
    }
    seen := map[string]struct{}{}
    var out []string

    var walk func(v interface{})
    walk = func(v interface{}) {
        switch t := v.(type) {
        case map[string]interface{}:
            for k, vv := range t {
                // интересуют ключи, похожие на id/fileId, но также проверим все строки на UUID
                if s, ok := vv.(string); ok {
                    if uuidRe.MatchString(s) {
                        if _, ok := seen[s]; !ok {
                            seen[s] = struct{}{}
                            out = append(out, s)
                        }
                    }
                }
                _ = k
                walk(vv)
            }
        case []interface{}:
            for _, it := range t {
                walk(it)
            }
        }
    }
    walk(any)
    return out, nil
}


