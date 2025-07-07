package ManyChat

import (
	"fmt"

	"tp-go-service/modules"
	"tp-go-service/modules/WeGoTrip"
)

const (
	ActionSetFieldValue = "set_field_value"
)

const (
	FieldAffiliateLink = "Ответ API URLs: афф.ссылка"
	FieldStatus        = "Ответ API URLs: status"
	FieldErrorMessage  = "Ответ API URLs: error_message"
	FieldErrorCode     = "Ответ API URLs: error_code"

	FieldTopPrice = "Ответ TOP-подборок [%d]: Price"
	FieldTopURL   = "Ответ TOP-подборок [%d]: URL"
	FieldTopImage = "Ответ TOP-подборок [%d]: Картинка"
	FieldTopTitle = "Ответ TOP-подборок [%d]: Название"
)

type ManyChat struct {
	version string
	content string
}

type Response struct {
	Version string  `json:"version"`
	Content Content `json:"content"`
}

type Content struct {
	Type     string   `json:"type"`
	Messages []string `json:"messages"`
	Actions  []Action `json:"actions"`
}

type Action struct {
	Action    string `json:"action"`
	FieldName string `json:"field_name"`
	Value     any    `json:"value"`
}

func New() *ManyChat {
	return &ManyChat{
		version: "v2",
		content: "instagram",
	}
}

func NewWithParams(content string) *ManyChat {
	return &ManyChat{
		version: "v2",
		content: content,
	}
}

func (mc *ManyChat) FromTravelPayoutsResponse(link string) Response {
	return Response{
		Version: mc.version,
		Content: Content{
			Type:     mc.content,
			Messages: []string{},
			Actions: []Action{
				{
					Action:    ActionSetFieldValue,
					FieldName: FieldAffiliateLink,
					Value:     link,
				},
				{
					Action:    ActionSetFieldValue,
					FieldName: FieldStatus,
					Value:     true,
				},
			},
		},
	}
}

func (mc *ManyChat) FromWeGoGetRespose(feedItems []WeGoTrip.FeedItem) Response {
	var actions []Action

	for i, item := range feedItems {
		index := i + 1

		actions = append(actions,
			Action{
				Action:    ActionSetFieldValue,
				FieldName: fmt.Sprintf(FieldTopPrice, index),
				Value:     item.Price,
			},
			Action{
				Action:    ActionSetFieldValue,
				FieldName: fmt.Sprintf(FieldTopURL, index),
				Value:     item.Link,
			},
			Action{
				Action:    ActionSetFieldValue,
				FieldName: fmt.Sprintf(FieldTopImage, index),
				Value:     item.Cover,
			},
			Action{
				Action:    ActionSetFieldValue,
				FieldName: fmt.Sprintf(FieldTopTitle, index),
				Value:     item.Title,
			},
		)
	}

	actions = append(actions,
		Action{
			Action:    ActionSetFieldValue,
			FieldName: FieldStatus,
			Value:     true,
		},
	)

	return Response{
		Version: mc.version,
		Content: Content{
			Type:     mc.content,
			Messages: []string{},
			Actions:  actions,
		},
	}
}

func (mc *ManyChat) FromError(err modules.APIError) Response {
	return Response{
		Version: mc.version,
		Content: Content{
			Type:     mc.content,
			Messages: []string{},
			Actions: []Action{
				{
					Action:    ActionSetFieldValue,
					FieldName: FieldErrorMessage,
					Value:     err.GetMessage(),
				},
				{
					Action:    ActionSetFieldValue,
					FieldName: FieldErrorCode,
					Value:     err.GetCode(),
				},
				{
					Action:    ActionSetFieldValue,
					FieldName: FieldStatus,
					Value:     false,
				},
			},
		},
	}
}

func (mc *ManyChat) FromValidationError(message string) Response {
	validationError := modules.NewError("invalid_request", message)
	return mc.FromError(validationError)
}
