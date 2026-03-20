package severavto

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

var (
	// ErrRateLimitExceeded возвращается когда запрос сделан чаще раз в 10 минут
	ErrRateLimitExceeded = errors.New("rate limit exceeded: wait 10 minutes between requests")

	// ErrNotModified возвращается когда данные не изменились с прошлого запроса (HTTP 304)
	ErrNotModified = errors.New("data not modified since last fetch")
)

// Client представляет HTTP клиент для API Северавто
type Client struct {
	baseURL      string
	apiKey       string
	httpClient   *http.Client
	logger       *slog.Logger

	// Rate limiting (10 минут между запросами)
	lastFetchTyres time.Time
	lastFetchRims  time.Time

	// HTTP caching
	lastModifiedTyres time.Time
	lastModifiedRims  time.Time
}

// ClientConfig содержит конфигурацию для HTTP клиента
type ClientConfig struct {
	BaseURL string        // https://webmim.svrauto.ru
	APIKey  string        // Уникальный ключ выгрузки
	Timeout time.Duration // Timeout для HTTP запросов
	Logger  *slog.Logger
}

// NewClient создает новый HTTP клиент для Северавто API
func NewClient(cfg ClientConfig) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}

	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &Client{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger: cfg.Logger.With("component", "severavto_client"),
	}
}

// FetchTyres скачивает XML с данными о шинах
func (c *Client) FetchTyres(ctx context.Context) (*CommoditiesXML, error) {
	return c.fetchXML(ctx, "tyre", &c.lastFetchTyres, &c.lastModifiedTyres)
}

// FetchRims скачивает XML с данными о дисках
func (c *Client) FetchRims(ctx context.Context) (*CommoditiesXML, error) {
	return c.fetchXML(ctx, "disc", &c.lastFetchRims, &c.lastModifiedRims)
}

// fetchXML скачивает и парсит XML для указанного типа товара
func (c *Client) fetchXML(
	ctx context.Context,
	productType string,
	lastFetch *time.Time,
	lastModified *time.Time,
) (*CommoditiesXML, error) {
	logger := c.logger.With("product_type", productType)

	// 1. Проверка rate limit (10 минут между запросами)
	if !lastFetch.IsZero() && time.Since(*lastFetch) < 10*time.Minute {
		remaining := 10*time.Minute - time.Since(*lastFetch)
		logger.Warn("Rate limit exceeded",
			"last_fetch", lastFetch,
			"wait_seconds", int(remaining.Seconds()),
		)
		return nil, ErrRateLimitExceeded
	}

	// 2. Построение URL (новый API v1)
	// Формат: https://webmim.svrauto.ru/api/v1/catalog/unload?access-token={API_KEY}&format=xml
	// ВАЖНО: API работает по HTTPS (перенаправляет с HTTP на HTTPS)
	url := fmt.Sprintf("%s/api/v1/catalog/unload?access-token=%s&format=xml",
		c.baseURL, c.apiKey)

	logger.Info("Fetching XML from Severavto API",
		"url", c.baseURL+"/api/v1/catalog/unload?access-token=***&format=xml",
		"base_url", c.baseURL,
		"api_key_length", len(c.apiKey))

	// 3. Создание HTTP запроса
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// 4. Добавление заголовка If-Modified-Since для кэширования
	if !lastModified.IsZero() {
		req.Header.Set("If-Modified-Since", lastModified.Format(http.TimeFormat))
		logger.Debug("Using cache header", "if_modified_since", lastModified.Format(http.TimeFormat))
	}

	// 5. Выполнение HTTP запроса
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	logger.Info("HTTP response received",
		"status_code", resp.StatusCode,
		"duration_ms", duration.Milliseconds(),
	)

	// 6. Обработка HTTP 304 (Not Modified)
	if resp.StatusCode == http.StatusNotModified {
		logger.Info("Data not modified since last fetch (HTTP 304)")
		return nil, ErrNotModified
	}

	// 7. Обработка HTTP 429 (Too Many Requests)
	if resp.StatusCode == http.StatusTooManyRequests {
		logger.Warn("Rate limit exceeded by server (HTTP 429)")
		return nil, ErrRateLimitExceeded
	}

	// 8. Проверка статуса ответа
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// 9. Парсинг XML (корневой элемент CATALOG)
	var catalog CatalogXML
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&catalog); err != nil {
		return nil, fmt.Errorf("xml decode: %w", err)
	}

	// 10. Фильтрация нужной категории по productType
	// tyre -> TYRE (ID=1), disc -> DISK (ID=2)
	var targetName string
	switch productType {
	case "tyre":
		targetName = "TYRE"
	case "disc":
		targetName = "DISK"
	default:
		return nil, fmt.Errorf("unknown product type: %s", productType)
	}

	var result *CommoditiesXML
	for i := range catalog.Commodities {
		if catalog.Commodities[i].Name == targetName {
			result = &catalog.Commodities[i]
			break
		}
	}

	if result == nil {
		return nil, fmt.Errorf("category %s not found in response", targetName)
	}

	logger.Info("XML parsed successfully",
		"category", result.Value,
		"target_name", targetName,
		"commodities_count", len(result.Commodities))

	// 11. Сохранение Last-Modified для кэширования
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		if parsed, err := time.Parse(http.TimeFormat, lm); err == nil {
			*lastModified = parsed
			logger.Debug("Updated Last-Modified", "last_modified", parsed)
		}
	}

	// 12. Обновление времени последнего запроса
	*lastFetch = time.Now()

	return result, nil
}
