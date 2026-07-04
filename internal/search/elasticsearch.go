package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

type SearchParams struct {
	Query    string
	MinPrice float64
	MaxPrice float64
	From     int
	Size     int
}

type SearchResult struct {
	Total int          `json:"total"`
	Hits  []ProductHit `json:"hits"`
}

type ProductHit struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Price       float64             `json:"price"`
	Stock       int                 `json:"stock"`
	ShopID      string              `json:"shop_id"`
	Status      string              `json:"status"`
	Score       float64             `json:"score"`
	Highlight   map[string][]string `json:"highlight,omitempty"`
}

type esSearchResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Score     float64             `json:"_score"`
			Source    map[string]any      `json:"_source"`
			Highlight map[string][]string `json:"highlight"`
		} `json:"hits"`
	}
}

func (c *Client) SearchProducts(ctx context.Context, params SearchParams) (*SearchResult, error) {
	if params.Size == 0 {
		params.Size = 10
	}

	query := buildQuery(params)

	body, err := json.Marshal(query)

	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/products/_search", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var esResp esSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&esResp); err != nil {
		return nil, err
	}

	return toSearchResult(esResp), nil
}

func buildQuery(params SearchParams) map[string]any {
	must := []map[string]any{}

	if params.Query != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":     params.Query,
				"fields":    []string{"name", "description"},
				"fuzziness": "AUTO",
			},
		})
	}

	// only display product with status "active"
	must = append(must, map[string]any{
		"term": map[string]any{"status": "active"},
	})

	// filter price if have
	filter := []map[string]any{}
	if params.MinPrice > 0 || params.MaxPrice > 0 {
		priceRange := map[string]any{}
		if params.MinPrice > 0 {
			priceRange["gte"] = params.MinPrice
		}

		if params.MaxPrice > 0 {
			priceRange["lte"] = params.MaxPrice
		}

		filter = append(filter, map[string]any{
			"range": map[string]any{"price": priceRange},
		})
	}

	return map[string]any{
		"from": params.From,
		"size": params.Size,
		"query": map[string]any{
			"bool": map[string]any{
				"must":   must,
				"filter": filter,
			},
		},
		"highlight": map[string]any{
			"fields": map[string]any{
				"name":        map[string]any{},
				"description": map[string]any{},
			},
		},
	}
}

func toSearchResult(esResp esSearchResponse) *SearchResult {
	hits := make([]ProductHit, 0, len(esResp.Hits.Hits))
	for _, h := range esResp.Hits.Hits {
		hit := ProductHit{
			Score: h.Score,
		}
		if v, ok := h.Source["id"].(string); ok {
			hit.ID = v
		}
		if v, ok := h.Source["name"].(string); ok {
			hit.Name = v
		}
		if v, ok := h.Source["description"].(string); ok {
			hit.Description = v
		}
		if v, ok := h.Source["price"].(float64); ok {
			hit.Price = v
		}
		if v, ok := h.Source["stock"].(float64); ok {
			hit.Stock = int(v)
		}
		if v, ok := h.Source["shop_id"].(string); ok {
			hit.ShopID = v
		}
		if v, ok := h.Source["status"].(string); ok {
			hit.Status = v
		}
		if h.Highlight != nil {
			hit.Highlight = h.Highlight
		}
		hits = append(hits, hit)
	}
	return &SearchResult{
		Total: esResp.Hits.Total.Value,
		Hits:  hits,
	}
}
