package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ConfluencePage struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	SpaceID string `json:"spaceId"`
	Body    struct {
		Storage struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body"`
	Version struct {
		Number int `json:"number"`
	} `json:"version"`
}

type ConfluenceClient struct {
	baseURL     string
	email       string
	apiToken    string
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

func NewConfluenceClient(baseURL, email, apiToken string) *ConfluenceClient {
	return &ConfluenceClient{
		baseURL:  baseURL,
		email:    email,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: NewRateLimiter(100 * time.Millisecond),
	}
}

func (c *ConfluenceClient) doRequest(method, path string, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+"/api/v2"+path, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.rateLimiter.Wait()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("invalid Confluence credentials (401)")
	}

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("%w: %s", errNotFound, string(respBody))
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

var errNotFound = fmt.Errorf("page not found")

// shouldRetryConfluenceWrite is true for rate limits, server errors, and transport failures.
// Client errors such as 409 Conflict are not retried with the same payload.
func shouldRetryConfluenceWrite(err error) bool {
	if err == nil {
		return false
	}
	var code int
	if n, _ := fmt.Sscanf(err.Error(), "API error %d:", &code); n == 1 {
		return code == 429 || code >= 500
	}
	msg := err.Error()
	return strings.Contains(msg, "making request:") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "EOF")
}

func (c *ConfluenceClient) GetPage(pageID string) (*ConfluencePage, error) {
	var result struct {
		ConfluencePage
	}

	// include-version ensures the current version number is present (listing APIs often omit it).
	path := "/pages/" + pageID + "?include-version=true"

	err := retryWithBackoff(3, 1*time.Second, func() error {
		data, err := c.doRequest("GET", path, nil)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, &result)
	})

	if errors.Is(err, errNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	page := result.ConfluencePage
	return &page, nil
}

func (c *ConfluenceClient) CreatePage(spaceID, title, body, parentID string) (*ConfluencePage, error) {
	payload := map[string]any{
		"spaceId": spaceID,
		"title":   title,
		"status":  "current",
		"body": map[string]any{
			"storage": map[string]string{
				"representation": "storage",
				"value":          body,
			},
		},
	}

	if parentID != "" {
		payload["parentId"] = parentID
	}

	var result ConfluencePage

	err := retryWithBackoff(3, 1*time.Second, func() error {
		data, err := c.doRequest("POST", "/pages", payload)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, &result)
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *ConfluenceClient) UpdatePage(pageID, title, body string, version int) (*ConfluencePage, error) {
	payload := map[string]any{
		"id":     pageID,
		"status": "current",
		"title":  title,
		"body": map[string]any{
			"storage": map[string]string{
				"representation": "storage",
				"value":          body,
			},
		},
		"version": map[string]any{
			"number":  version + 1,
			"message": "Updated by doc-helper",
		},
	}

	var result ConfluencePage
	const maxRetries = 3
	baseDelay := 1 * time.Second
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		data, err := c.doRequest("PUT", "/pages/"+pageID, payload)
		if err == nil {
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, err
			}
			return &result, nil
		}
		lastErr = err
		if attempt < maxRetries && shouldRetryConfluenceWrite(err) {
			delay := baseDelay * time.Duration(1<<uint(attempt))
			fmt.Fprintf(os.Stderr, "  Retry %d/%d after %v: %v\n", attempt+1, maxRetries, delay, err)
			time.Sleep(delay)
			continue
		}
		break
	}

	return nil, lastErr
}

func (c *ConfluenceClient) GetChildPages(parentID string) ([]ConfluencePage, error) {
	var allPages []ConfluencePage
	path := "/pages/" + parentID + "/children"

	for path != "" {
		var result struct {
			Results []ConfluencePage `json:"results"`
			Links   struct {
				Next string `json:"next"`
			} `json:"_links"`
		}

		err := retryWithBackoff(3, 1*time.Second, func() error {
			data, err := c.doRequest("GET", path, nil)
			if err != nil {
				return err
			}
			return json.Unmarshal(data, &result)
		})

		if err != nil {
			return nil, err
		}

		allPages = append(allPages, result.Results...)

		if result.Links.Next != "" {
			path = resolveNextURL(result.Links.Next)
		} else {
			path = ""
		}
	}

	return allPages, nil
}

func resolveNextURL(next string) string {
	if next == "" {
		return ""
	}
	// The Confluence v2 API returns paths like /api/v2/pages?cursor=xxx
	// We only need the path portion after /api/v2
	if len(next) > 8 && next[:8] == "/api/v2/" {
		return next[7:] // strip "/api/v2" prefix, keep the rest
	}
	return next
}
