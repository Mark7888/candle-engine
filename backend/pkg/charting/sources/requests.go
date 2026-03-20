package sources

import (
	"encoding/json"
	"io"
	"net/http"
)

func GetRequestParse[T any](url string, params map[string]string, jsonPath ...string) (v *T, statusCode int, err error) {
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

	// If jsonPath is provided, navigate to the nested field
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, -4, err
	}

	if len(jsonPath) > 0 {
		for _, key := range jsonPath {
			if m, ok := data.(map[string]interface{}); ok {
				data = m[key]
			} else {
				return nil, -5, err
			}
		}
		// Re-marshal the nested data for final unmarshal
		body, err = json.Marshal(data)
		if err != nil {
			return nil, -6, err
		}
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, -4, err
	}
	return &result, resp.StatusCode, nil
}
