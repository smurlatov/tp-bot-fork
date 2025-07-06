package WeGoTrip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"tp-go-service/modules"
)

type WeGoTrip struct {
	client *http.Client
}

type WeGoTripError struct {
	modules.BaseError
}

func NewWeGoTripError(code, message string) modules.APIError {
	return &WeGoTripError{
		BaseError: modules.BaseError{
			Code:    code,
			Message: message,
		},
	}
}

type WeGoTripProduct struct {
	ID    int          `json:"id"`
	Title string       `json:"title"`
	Slug  string       `json:"slug"`
	Cover string       `json:"cover"`
	Price float64      `json:"price"`
	City  WeGoTripCity `json:"city"`
}

type WeGoTripCity struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type WeGoTripResponse struct {
	Data WeGoTripData `json:"data"`
}

type WeGoTripData struct {
	Count    int               `json:"count"`
	Pages    int               `json:"pages"`
	Current  int               `json:"current"`
	Results  []WeGoTripProduct `json:"results"`
	MaxPrice float64           `json:"maxPrice"`
}

type WeGoTripErrorStruct struct {
	Message string `json:"message"`
}

type WeGoTripErrorResponse struct {
	Errors []WeGoTripErrorStruct `json:"errors"`
}

type FeedItem struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	Slug     string  `json:"slug"`
	CitySlug string  `json:"city_slug"`
	Price    float64 `json:"price"`
	Cover    string  `json:"cover"`
	Link     string  `json:"link"`
}

func New() *WeGoTrip {
	return &WeGoTrip{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (wg *WeGoTrip) GetFeed(city, lang, currency string, page int) ([]FeedItem, modules.APIError) {
	if lang == "" {
		lang = "RU"
	}
	if currency == "" {
		currency = "RUB"
	}
	if page <= 0 {
		page = 1
	}

	cityID := 0
	domain := ""

	if cityID = GetCOMWeGoTripCityID(city); cityID != 0 {
		domain = "com"
	} else if cityID = GetRUWeGoTripCityID(city); cityID != 0 {
		domain = "ru"
	}

	if cityID == 0 {
		return nil, NewWeGoTripError("city_not_found", "нет такого города")
	}

	var baseURL string
	if domain == "ru" {
		baseURL = "https://wegotrip.ru"
	} else {
		baseURL = "https://app.wegotrip.com"
	}

	requestURL := fmt.Sprintf("%s/api/v2/products/popular/?city=%d&lang=%s&currency=%s",
		baseURL, cityID, strings.ToLower(lang), currency)

	resp, err := wg.client.Get(requestURL)
	if err != nil {
		return nil, NewWeGoTripError("network_error", "ошибка запроса к WeGoTrip API")
	}
	defer resp.Body.Close()

	var responseBody bytes.Buffer
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		return nil, NewWeGoTripError("response_error", "ошибка чтения ответа")
	}

	responseBytes := responseBody.Bytes()

	var errorResponse WeGoTripErrorResponse
	if err := json.Unmarshal(responseBytes, &errorResponse); err == nil && len(errorResponse.Errors) > 0 {
		errorMessage := errorResponse.Errors[0].Message
		return nil, NewWeGoTripError("wegotrip_api_error", errorMessage)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, NewWeGoTripError("api_error", fmt.Sprintf("API вернул ошибку %d", resp.StatusCode))
	}

	var apiResponse WeGoTripResponse
	if err := json.Unmarshal(responseBytes, &apiResponse); err != nil {
		return nil, NewWeGoTripError("parse_error", "ошибка парсинга ответа")
	}

	results := apiResponse.Data.Results
	totalItems := len(results)

	startIndex := (page - 1) * 3
	endIndex := startIndex + 3

	if startIndex >= totalItems {
		return []FeedItem{}, nil
	}

	if endIndex > totalItems {
		endIndex = totalItems
	}

	paginatedResults := results[startIndex:endIndex]

	var feedItems []FeedItem
	for _, product := range paginatedResults {

		var linkDomain string
		if domain == "ru" {
			linkDomain = "wegotrip.ru"
		} else {
			linkDomain = "app.wegotrip.com"
		}

		link := fmt.Sprintf("https://%s/%s-d%d/%s-p%d",
			linkDomain, product.City.Slug, cityID, product.Slug, product.ID)

		feedItem := FeedItem{
			ID:       product.ID,
			Title:    product.Title,
			Slug:     product.Slug,
			CitySlug: product.City.Slug,
			Price:    product.Price,
			Cover:    product.Cover,
			Link:     link,
		}

		feedItems = append(feedItems, feedItem)
	}

	return feedItems, nil
}
