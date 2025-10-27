package sources

import (
	"encoding/json"
	"io"
	"net/http"
)

func GetRequestParse[T any](url string, params map[string]string) (v *T, statusCode int, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, -1, err
	}

	q := req.URL.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, -2, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, -3, err
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, -4, err
	}
	return &result, resp.StatusCode, nil
}
