package TravelPayouts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"tp-go-service/modules"
)

type TravelPayouts struct {
	token  string
	trs    int
	marker int
	client *http.Client
}

type TravelPayoutsError struct {
	modules.BaseError
}

func NewTravelPayoutsError(code, message string) modules.APIError {
	return &TravelPayoutsError{
		BaseError: modules.BaseError{
			Code:    code,
			Message: message,
		},
	}
}

type TravelPayoutsRequest struct {
	TRS     int                     `json:"trs"`
	Marker  int                     `json:"marker"`
	Shorten bool                    `json:"shorten"`
	Links   []TravelPayoutsLinkItem `json:"links"`
}

type TravelPayoutsLinkItem struct {
	URL   string `json:"url"`
	SubID string `json:"sub_id"`
}

type TravelPayoutsResponse struct {
	Result TravelPayoutsResult `json:"result"`
	Code   string              `json:"code"`
	Status int                 `json:"status"`
}

type TravelPayoutsResult struct {
	TRS     int                         `json:"trs"`
	Marker  int                         `json:"marker"`
	Shorten bool                        `json:"shorten"`
	Links   []TravelPayoutsResponseLink `json:"links"`
}

type TravelPayoutsResponseLink struct {
	URL        string `json:"url"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	PartnerURL string `json:"partner_url"`
}

type TravelPayoutsErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func New(token, trs, marker string) (*TravelPayouts, modules.APIError) {
	trsInt, err := strconv.Atoi(trs)
	if err != nil {
		return nil, NewTravelPayoutsError("invalid_trs", "неверный формат TRS")
	}

	markerInt, err := strconv.Atoi(marker)
	if err != nil {
		return nil, NewTravelPayoutsError("invalid_marker", "неверный формат Marker")
	}

	return &TravelPayouts{
		token:  token,
		trs:    trsInt,
		marker: markerInt,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GetFromLink создает аффилиатную ссылку из обычной ссылки
func (tp *TravelPayouts) GetFromLink(originalLink string) (string, modules.APIError) {
	// Создаем запрос к Travelpayouts API
	request := TravelPayoutsRequest{
		TRS:     tp.trs,
		Marker:  tp.marker,
		Shorten: true,
		Links: []TravelPayoutsLinkItem{
			{
				URL:   originalLink,
				SubID: "social_tool_main",
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", NewTravelPayoutsError("json_error", "ошибка сериализации данных")
	}

	req, err := http.NewRequest("POST", "https://api.travelpayouts.com/links/v1/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", NewTravelPayoutsError("request_error", "ошибка создания запроса")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Access-Token", tp.token)

	resp, err := tp.client.Do(req)
	if err != nil {
		return "", NewTravelPayoutsError("network_error", "ошибка запроса к Travelpayouts API")
	}
	defer resp.Body.Close()

	var responseBody bytes.Buffer
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		return "", NewTravelPayoutsError("response_error", "ошибка чтения ответа")
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp TravelPayoutsErrorResponse
		if err := json.Unmarshal(responseBody.Bytes(), &errorResp); err != nil {
			return "", NewTravelPayoutsError("api_error", fmt.Sprintf("API вернул ошибку %d", resp.StatusCode))
		}

		if errorResp.Error != "" {
			return "", NewTravelPayoutsError(errorResp.Code, errorResp.Error)
		}
		return "", NewTravelPayoutsError(errorResp.Code, errorResp.Message)
	}

	var apiResponse TravelPayoutsResponse
	if err := json.Unmarshal(responseBody.Bytes(), &apiResponse); err != nil {
		return "", NewTravelPayoutsError("parse_error", "ошибка парсинга ответа")
	}

	if apiResponse.Code != "success" {
		return "", NewTravelPayoutsError(apiResponse.Code, "API вернул код ошибки")
	}

	if len(apiResponse.Result.Links) == 0 {
		return "", NewTravelPayoutsError("no_links", "API не вернул ссылок")
	}

	link := apiResponse.Result.Links[0]

	if link.Code != "success" {
		return "", NewTravelPayoutsError(link.Code, link.Message)
	}

	if link.PartnerURL == "" {
		return "", NewTravelPayoutsError("empty_partner_url", "API не вернул партнерскую ссылку")
	}

	return link.PartnerURL, nil
}
