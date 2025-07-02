package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TravelPayoutsResponse представляет ответ от Travelpayouts API
type TravelPayoutsResponse struct {
	Link  string `json:"link"`
	Error string `json:"error"`
}

// makeAffiliateLink создает аффилиатную ссылку через Travelpayouts API
func makeAffiliateLink(originalLink, token, trs, marker string) (string, error) {
	logger.WithFields(map[string]interface{}{
		"original_link": originalLink,
		"token":         token,
		"trs":           trs,
		"marker":        marker,
	}).Info("Создание аффилиатной ссылки")

	// Базовый URL API Travelpayouts для создания партнерских ссылок
	// Поскольку конкретного API для создания партнерских ссылок в документации 
	// не найдено, используем общий подход с добавлением параметров
	
	// Для демонстрации и тестирования создадим простую логику
	// которая добавляет партнерские параметры к ссылке
	affiliateLink, err := buildAffiliateLink(originalLink, token, trs, marker)
	if err != nil {
		return "", fmt.Errorf("ошибка построения аффилиатной ссылки: %w", err)
	}

	// В реальном случае здесь был бы HTTP запрос к Travelpayouts API
	// Пример:
	// response, err := callTravelPayoutsAPI(originalLink, token, trs, marker)
	// if err != nil {
	//     return "", err
	// }
	// return response.Link, nil

	return affiliateLink, nil
}

// buildAffiliateLink строит аффилиатную ссылку добавляя необходимые параметры
func buildAffiliateLink(originalLink, token, trs, marker string) (string, error) {
	parsedURL, err := url.Parse(originalLink)
	if err != nil {
		return "", fmt.Errorf("некорректная ссылка: %w", err)
	}

	// Добавляем партнерские параметры
	values := parsedURL.Query()
	
	// Обязательный параметр sub_id согласно требованиям
	values.Set("sub_id", "social_tool_main")
	
	// Добавляем переданные параметры
	values.Set("token", token)
	values.Set("trs", trs)
	values.Set("marker", marker)
	
	// Добавляем utm метки для отслеживания
	values.Set("utm_source", "travelpayouts")
	values.Set("utm_medium", "affiliate")
	values.Set("utm_campaign", "social_tool")

	parsedURL.RawQuery = values.Encode()
	
	return parsedURL.String(), nil
}

// callTravelPayoutsAPI делает реальный запрос к Travelpayouts API
// Эта функция будет использоваться когда будет найдена точная документация по API
func callTravelPayoutsAPI(originalLink, token, trs, marker string) (*TravelPayoutsResponse, error) {
	// Пример реального запроса к API (когда будет доступна точная документация)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Формируем запрос к API Travelpayouts
	// URL API нужно уточнить в документации
	apiURL := "https://api.travelpayouts.com/v1/affiliate_links"
	
	payload := map[string]string{
		"url":     originalLink,
		"token":   token,
		"trs":     trs,
		"marker":  marker,
		"sub_id":  "social_tool_main",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Access-Token", token)

	logger.WithFields(map[string]interface{}{
		"url":     apiURL,
		"payload": string(payloadBytes),
	}).Info("Отправка запроса к Travelpayouts API")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API вернул статус %d", resp.StatusCode)
	}

	var response TravelPayoutsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf("ошибка API: %s", response.Error)
	}

	logger.WithField("affiliate_link", response.Link).Info("Получена аффилиатная ссылка от API")
	
	return &response, nil
} 