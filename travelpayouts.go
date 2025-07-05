package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"strings"
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
	Message    string `json:"message"`
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
func makeAffiliateLink(originalLink, token, trs, marker string) (string, *TPError, error) {
	logger.WithFields(map[string]interface{}{
		"original_link": originalLink,
		"token":         token,
		"trs":           trs,
		"marker":        marker,
	}).Info("Создание аффилиатной ссылки через Travelpayouts API")

	// Конвертируем строки в числа
	trsInt, err := strconv.Atoi(trs)
	if err != nil {
		return "", &TPError{Code: "invalid_trs", Message: "неверный формат TRS"}, fmt.Errorf("неверный формат TRS: %v", err)
	}

	markerInt, err := strconv.Atoi(marker)
	if err != nil {
		return "", &TPError{Code: "invalid_marker", Message: "неверный формат Marker"}, fmt.Errorf("неверный формат Marker: %v", err)
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
		return "", &TPError{Code: "json_error", Message: "ошибка сериализации данных"}, fmt.Errorf("ошибка сериализации данных: %v", err)
	}

	logger.WithField("request_body", string(jsonData)).Info("Отправка запроса к Travelpayouts API")

	// Создаем HTTP запрос
	req, err := http.NewRequest("POST", "https://api.travelpayouts.com/links/v1/create", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.WithError(err).Error("Ошибка создания HTTP запроса")
		return "", &TPError{Code: "request_error", Message: "ошибка создания запроса"}, fmt.Errorf("ошибка создания запроса: %v", err)
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
		return "", &TPError{Code: "network_error", Message: "ошибка запроса к Travelpayouts API"}, fmt.Errorf("ошибка запроса к Travelpayouts API: %v", err)
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
		return "", &TPError{Code: "response_error", Message: "ошибка чтения ответа"}, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	responseString := responseBody.String()
	logger.WithField("response_body", responseString).Info("Тело ответа от Travelpayouts API")

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		// Пытаемся парсить ошибку
		var errorResp TravelPayoutsErrorResponse
		if err := json.Unmarshal(responseBody.Bytes(), &errorResp); err != nil {
			logger.WithError(err).Error("Ошибка парсинга ошибки от API")
			return "", &TPError{Code: "api_error", Message: fmt.Sprintf("API вернул ошибку %d", resp.StatusCode)}, fmt.Errorf("API вернул ошибку %d: %s", resp.StatusCode, responseString)
		}
		
		logger.WithFields(map[string]interface{}{
			"error_code":    errorResp.Code,
			"error_message": errorResp.Error,
			"status":        errorResp.Status,
		}).Error("Travelpayouts API вернул ошибку")
		
		if errorResp.Error != "" {
			return "", &TPError{Code: errorResp.Code, Message: errorResp.Error}, fmt.Errorf("ошибка Travelpayouts API: %s", errorResp.Error)
		}
		return "", &TPError{Code: errorResp.Code, Message: errorResp.Message}, fmt.Errorf("ошибка Travelpayouts API %d: %s", resp.StatusCode, errorResp.Message)
	}

	// Парсим успешный ответ
	var apiResponse TravelPayoutsResponse
	if err := json.Unmarshal(responseBody.Bytes(), &apiResponse); err != nil {
		logger.WithError(err).Error("Ошибка парсинга ответа API")
		return "", &TPError{Code: "parse_error", Message: "ошибка парсинга ответа"}, fmt.Errorf("ошибка парсинга ответа: %v", err)
	}

	// Проверяем код ответа
	if apiResponse.Code != "success" {
		logger.WithField("response_code", apiResponse.Code).Error("API вернул неуспешный код")
		return "", &TPError{Code: apiResponse.Code, Message: "API вернул код ошибки"}, fmt.Errorf("API вернул код ошибки: %s", apiResponse.Code)
	}

	// Проверяем наличие ссылок
	if len(apiResponse.Result.Links) == 0 {
		logger.Error("В ответе API нет ссылок")
		return "", &TPError{Code: "no_links", Message: "API не вернул ссылок"}, fmt.Errorf("API не вернул ссылок")
	}

	// Получаем первую ссылку
	link := apiResponse.Result.Links[0]
	
	// Проверяем статус ссылки - КЛЮЧЕВАЯ ЛОГИКА ДЛЯ ОБРАБОТКИ ОШИБОК ОТ TP
	if link.Code != "success" {
		logger.WithFields(map[string]interface{}{
			"link_code":    link.Code,
			"link_message": link.Message,
		}).Error("Ссылка имеет статус ошибки")
		
		// Возвращаем ошибку от TP API с их кодом и сообщением
		return "", &TPError{Code: link.Code, Message: link.Message}, fmt.Errorf("ошибка создания ссылки: %s - %s", link.Code, link.Message)
	}

	// Проверяем наличие партнерской ссылки
	if link.PartnerURL == "" {
		logger.Error("Партнерская ссылка пуста")
		return "", &TPError{Code: "empty_partner_url", Message: "API не вернул партнерскую ссылку"}, fmt.Errorf("API не вернул партнерскую ссылку")
	}

	logger.WithFields(map[string]interface{}{
		"original_url":  link.URL,
		"partner_url":   link.PartnerURL,
		"response_code": link.Code,
	}).Info("Успешно создана аффилиатная ссылка")

	return link.PartnerURL, nil, nil
}

// buildAffiliateLink - старая функция для тестирования (оставляем как fallback)
func buildAffiliateLink(originalLink, token, trs, marker string) (string, error) {
	logger.Warn("Используется fallback метод создания ссылки (не реальный API)")
	
	// Простое добавление параметров к URL (для тестирования)
	return fmt.Sprintf("%s?sub_id=social_tool_main&token=%s&trs=%s&marker=%s", 
		originalLink, token, trs, marker), nil
}

// getTravelpayoutsFeed получает фид поездок для указанного города
func getTravelpayoutsFeed(city, lang, currency string, page int) ([]FeedItem, *TPError, error) {
	// Устанавливаем значения по умолчанию
	if lang == "" {
		lang = "RU"
	}
	if currency == "" {
		currency = "RUB"
	}
	if page <= 0 {
		page = 1
	}
	
	logger.WithFields(map[string]interface{}{
		"city":     city,
		"lang":     lang,
		"currency": currency,
		"page":     page,
	}).Info("Поиск города в картах WeGoTrip")
	
	// Поиск города в картах (сначала COM, потом RU)
	cityID := 0
	domain := ""
	
	// Проверяем международную карту
	if cityID = GetCOMWeGoTripCityID(city); cityID != 0 {
		domain = "com"
		logger.WithFields(map[string]interface{}{
			"city":     city,
			"city_id":  cityID,
			"domain":   domain,
		}).Info("Город найден в международной карте")
	} else if cityID = GetRUWeGoTripCityID(city); cityID != 0 {
		domain = "ru"
		logger.WithFields(map[string]interface{}{
			"city":     city,
			"city_id":  cityID,
			"domain":   domain,
		}).Info("Город найден в русской карте")
	}
	
	// Если город не найден
	if cityID == 0 {
		logger.WithField("city", city).Error("Город не найден ни в одной из карт")
		return nil, &TPError{Code: "city_not_found", Message: "нет такого города"}, fmt.Errorf("город не найден: %s", city)
	}
	
	// Формируем URL для запроса к WeGoTrip API
	var baseURL string
	if domain == "ru" {
		baseURL = "https://wegotrip.ru"
	} else {
		baseURL = "https://app.wegotrip.com"
	}
	
	requestURL := fmt.Sprintf("%s/api/v2/products/popular/?city=%d&lang=%s&currency=%s", 
		baseURL, cityID, strings.ToLower(lang), currency)
	
	logger.WithFields(map[string]interface{}{
		"request_url": requestURL,
		"city_id":     cityID,
		"domain":      domain,
	}).Info("Отправка запроса к WeGoTrip API")
	
	// Делаем HTTP запрос
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Get(requestURL)
	if err != nil {
		logger.WithError(err).Error("Ошибка выполнения HTTP запроса к WeGoTrip")
		return nil, &TPError{Code: "network_error", Message: "ошибка запроса к WeGoTrip API"}, fmt.Errorf("ошибка запроса к WeGoTrip: %v", err)
	}
	defer resp.Body.Close()
	
	logger.WithFields(map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
	}).Info("Получен ответ от WeGoTrip API")
	
	// Читаем ответ
	var responseBody bytes.Buffer
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		logger.WithError(err).Error("Ошибка чтения ответа WeGoTrip")
		return nil, &TPError{Code: "response_error", Message: "ошибка чтения ответа"}, fmt.Errorf("ошибка чтения ответа: %v", err)
	}
	
	responseString := responseBody.String()
	logger.WithField("response_body", responseString).Info("Тело ответа от WeGoTrip API")
	
	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		logger.WithField("status_code", resp.StatusCode).Error("WeGoTrip API вернул ошибку")
		return nil, &TPError{Code: "api_error", Message: fmt.Sprintf("API вернул ошибку %d", resp.StatusCode)}, fmt.Errorf("API вернул ошибку %d", resp.StatusCode)
	}
	
	// Парсим ответ
	var apiResponse WeGoTripResponse
	if err := json.Unmarshal(responseBody.Bytes(), &apiResponse); err != nil {
		logger.WithError(err).Error("Ошибка парсинга ответа WeGoTrip API")
		return nil, &TPError{Code: "parse_error", Message: "ошибка парсинга ответа"}, fmt.Errorf("ошибка парсинга ответа: %v", err)
	}
	
	// Применяем пагинацию (3 элемента на страницу)
	results := apiResponse.Data.Results
	totalItems := len(results)
	
	// Вычисляем индексы для пагинации
	startIndex := (page - 1) * 3
	endIndex := startIndex + 3
	
	if startIndex >= totalItems {
		logger.WithFields(map[string]interface{}{
			"page":        page,
			"total_items": totalItems,
			"start_index": startIndex,
		}).Info("Запрошенная страница превышает количество доступных данных")
		return []FeedItem{}, nil, nil
	}
	
	if endIndex > totalItems {
		endIndex = totalItems
	}
	
	paginatedResults := results[startIndex:endIndex]
	
	logger.WithFields(map[string]interface{}{
		"total_items":      totalItems,
		"page":            page,
		"start_index":     startIndex,
		"end_index":       endIndex,
		"paginated_count": len(paginatedResults),
	}).Info("Применена пагинация")
	
	// Преобразуем в FeedItem с генерацией ссылок
	var feedItems []FeedItem
	for _, product := range paginatedResults {
		// Генерируем ссылку в формате: https://wegotrip.{domain}/${city-slug}-d${city-id}/${product-slug}-p${product-id}
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
		
		logger.WithFields(map[string]interface{}{
			"product_id":      product.ID,
			"product_title":   product.Title,
			"generated_link":  link,
		}).Info("Сгенерирована ссылка для продукта")
	}
	
	logger.WithField("feed_items_count", len(feedItems)).Info("Успешно создан фид поездок")
	
	return feedItems, nil, nil
} 