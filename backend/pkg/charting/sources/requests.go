package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetRequestParse[T any, E SourceError](url string, params map[string]string, jsonPath ...string) (v *T, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiError E
		if err := json.Unmarshal(body, &apiError); err != nil {
			return nil, fmt.Errorf("Error why parsing API error response: %w; original body: %s", err, string(body))
		}

		return nil, apiError.Error()
	}

	// If jsonPath is provided, navigate to the nested field
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	if len(jsonPath) > 0 {
		for _, key := range jsonPath {
			if m, ok := data.(map[string]interface{}); ok {
				data = m[key]
			} else {
				return nil, fmt.Errorf("invalid json path: %v", jsonPath)
			}
		}
		// Re-marshal the nested data for final unmarshal
		body, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("error re-marshaling nested data: %w", err)
		}
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling final result: %w", err)
	}
	return &result, nil
}
