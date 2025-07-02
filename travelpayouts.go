package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// TravelPayoutsRequest представляет запрос к Travelpayouts API
type TravelPayoutsRequest struct {
	TRS     int                      `json:"trs"`
	Marker  int                      `json:"marker"`
	Shorten bool                     `json:"shorten"`
	Links   []TravelPayoutsLinkItem  `json:"links"`
}

// TravelPayoutsLinkItem представляет элемент массива ссылок
type TravelPayoutsLinkItem struct {
	URL   string `json:"url"`
	SubID string `json:"sub_id"`
}

// TravelPayoutsResponse представляет ответ от Travelpayouts API
type TravelPayoutsResponse struct {
	Result TravelPayoutsResult `json:"result"`
	Code   string              `json:"code"`
	Status int                 `json:"status"`
}

// TravelPayoutsResult представляет результат в ответе
type TravelPayoutsResult struct {
	TRS     int                          `json:"trs"`
	Marker  int                          `json:"marker"`
	Shorten bool                         `json:"shorten"`
	Links   []TravelPayoutsResponseLink  `json:"links"`
}

// TravelPayoutsResponseLink представляет ссылку в результате
type TravelPayoutsResponseLink struct {
	URL        string `json:"url"`
	Code       string `json:"code"`
	PartnerURL string `json:"partner_url"`
}

// TravelPayoutsErrorResponse для обработки ошибок
type TravelPayoutsErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// makeAffiliateLink создает аффилиатную ссылку через Travelpayouts API
func makeAffiliateLink(originalLink, token, trs, marker string) (string, error) {
	logger.WithFields(map[string]interface{}{
		"original_link": originalLink,
		"token":         token,
		"trs":           trs,
		"marker":        marker,
	}).Info("Создание аффилиатной ссылки через Travelpayouts API")

	// Конвертируем строки в числа
	trsInt, err := strconv.Atoi(trs)
	if err != nil {
		return "", fmt.Errorf("неверный формат TRS: %v", err)
	}

	markerInt, err := strconv.Atoi(marker)
	if err != nil {
		return "", fmt.Errorf("неверный формат Marker: %v", err)
	}

	// Создаем запрос к Travelpayouts API
	request := TravelPayoutsRequest{
		TRS:     trsInt,
		Marker:  markerInt,
		Shorten: true,
		Links: []TravelPayoutsLinkItem{
			{
				URL:   originalLink,
				SubID: "social_tool_main",
			},
		},
	}

	// Конвертируем в JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		logger.WithError(err).Error("Ошибка сериализации JSON")
		return "", fmt.Errorf("ошибка сериализации данных: %v", err)
	}

	logger.WithField("request_body", string(jsonData)).Info("Отправка запроса к Travelpayouts API")

	// Создаем HTTP запрос
	req, err := http.NewRequest("POST", "https://api.travelpayouts.com/links/v1/create", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.WithError(err).Error("Ошибка создания HTTP запроса")
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Access-Token", token)

	// Выполняем запрос
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.WithError(err).Error("Ошибка выполнения HTTP запроса")
		return "", fmt.Errorf("ошибка запроса к Travelpayouts API: %v", err)
	}
	defer resp.Body.Close()

	logger.WithFields(map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
	}).Info("Получен ответ от Travelpayouts API")

	// Читаем ответ
	var responseBody bytes.Buffer
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		logger.WithError(err).Error("Ошибка чтения ответа")
		return "", fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	responseString := responseBody.String()
	logger.WithField("response_body", responseString).Info("Тело ответа от Travelpayouts API")

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		// Пытаемся парсить ошибку
		var errorResp TravelPayoutsErrorResponse
		if err := json.Unmarshal(responseBody.Bytes(), &errorResp); err != nil {
			logger.WithError(err).Error("Ошибка парсинга ошибки от API")
			return "", fmt.Errorf("API вернул ошибку %d: %s", resp.StatusCode, responseString)
		}
		
		logger.WithFields(map[string]interface{}{
			"error_code":    errorResp.Code,
			"error_message": errorResp.Error,
			"status":        errorResp.Status,
		}).Error("Travelpayouts API вернул ошибку")
		
		if errorResp.Error != "" {
			return "", fmt.Errorf("ошибка Travelpayouts API: %s", errorResp.Error)
		}
		return "", fmt.Errorf("ошибка Travelpayouts API %d: %s", resp.StatusCode, errorResp.Message)
	}

	// Парсим успешный ответ
	var apiResponse TravelPayoutsResponse
	if err := json.Unmarshal(responseBody.Bytes(), &apiResponse); err != nil {
		logger.WithError(err).Error("Ошибка парсинга ответа API")
		return "", fmt.Errorf("ошибка парсинга ответа: %v", err)
	}

	// Проверяем код ответа
	if apiResponse.Code != "success" {
		logger.WithField("response_code", apiResponse.Code).Error("API вернул неуспешный код")
		return "", fmt.Errorf("API вернул код ошибки: %s", apiResponse.Code)
	}

	// Проверяем наличие ссылок
	if len(apiResponse.Result.Links) == 0 {
		logger.Error("В ответе API нет ссылок")
		return "", fmt.Errorf("API не вернул ссылок")
	}

	// Получаем первую ссылку
	link := apiResponse.Result.Links[0]
	
	// Проверяем статус ссылки
	if link.Code != "success" {
		logger.WithField("link_code", link.Code).Error("Ссылка имеет статус ошибки")
		return "", fmt.Errorf("ошибка создания ссылки: %s", link.Code)
	}

	// Проверяем наличие партнерской ссылки
	if link.PartnerURL == "" {
		logger.Error("Партнерская ссылка пуста")
		return "", fmt.Errorf("API не вернул партнерскую ссылку")
	}

	logger.WithFields(map[string]interface{}{
		"original_url":  link.URL,
		"partner_url":   link.PartnerURL,
		"response_code": link.Code,
	}).Info("Успешно создана аффилиатная ссылка")

	return link.PartnerURL, nil
}

// buildAffiliateLink - старая функция для тестирования (оставляем как fallback)
func buildAffiliateLink(originalLink, token, trs, marker string) (string, error) {
	logger.Warn("Используется fallback метод создания ссылки (не реальный API)")
	
	// Простое добавление параметров к URL (для тестирования)
	return fmt.Sprintf("%s?sub_id=social_tool_main&token=%s&trs=%s&marker=%s", 
		originalLink, token, trs, marker), nil
} 