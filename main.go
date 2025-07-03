package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Brand struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	BrandName string `json:"brand_name" gorm:"unique;not null"`
	BrandLink string `json:"brand_link" gorm:"not null"`
}

type GetFromLinkRequest struct {
	Link   string `json:"link" binding:"required"`
	Token  string `json:"token" binding:"required"`
	TRS    string `json:"trs" binding:"required"`
	Marker string `json:"marker" binding:"required"`
}

type GetFromBrandRequest struct {
	BrandName string `json:"brand_name" binding:"required"`
	Token     string `json:"token" binding:"required"`
	TRS       string `json:"trs" binding:"required"`
	Marker    string `json:"marker" binding:"required"`
}

// Новые структуры для формата ответа v2
type V2Response struct {
	Version string    `json:"version"`
	Content V2Content `json:"content"`
}

type V2Content struct {
	Type     string      `json:"type"`
	Messages []string    `json:"messages"`
	Actions  []V2Action  `json:"actions"`
}

type V2Action struct {
	Action    string      `json:"action"`
	FieldName string      `json:"field_name"`
	Value     interface{} `json:"value"`
}

// Структура для ошибок от Travelpayouts API
type TPError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Link  string `json:"link,omitempty"`
	Error string `json:"error,omitempty"`
}

var db *gorm.DB
var logger *logrus.Logger

func main() {
	// Инициализация логгера
	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	
	// Инициализация базы данных
	var err error
	db, err = gorm.Open(sqlite.Open("brands.db"), &gorm.Config{})
	if err != nil {
		logger.Fatal("Не удалось подключиться к базе данных:", err)
	}
	
	// Автомиграция
	db.AutoMigrate(&Brand{})
	
	// Инициализация тестовых данных
	initTestData()
	
	// Настройка Gin в production режиме
	gin.SetMode(gin.ReleaseMode)
	
	r := gin.New()
	r.Use(gin.LoggerWithWriter(logger.Out), gin.Recovery())
	
	// Middleware для логирования
	r.Use(func(c *gin.Context) {
		logger.WithFields(logrus.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
		}).Info("HTTP Request")
		c.Next()
	})
	
	// Роуты
	api := r.Group("/api")
	{
		api.POST("/getFromLink", getFromLink)
		api.POST("/getFromBrand", getFromBrand)
	}
	
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	logger.Info("Сервер запущен на порту ", port)
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Ошибка запуска сервера:", err)
	}
}

func initTestData() {
	// Добавляем тестовые бренды
	brands := []Brand{
		{BrandName: "booking", BrandLink: "https://www.booking.com"},
		{BrandName: "agoda", BrandLink: "https://www.agoda.com"},
		{BrandName: "aviasales", BrandLink: "https://aviasales.com"},
		{BrandName: "hotels", BrandLink: "https://hotels.com"},
		{BrandName: "expedia", BrandLink: "https://www.expedia.com"},
	}
	
	for _, brand := range brands {
		db.Where(Brand{BrandName: brand.BrandName}).FirstOrCreate(&brand)
	}
	
	logger.Info("Тестовые данные инициализированы")
}

func getFromLink(c *gin.Context) {
	var req GetFromLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Ошибка валидации запроса getFromLink")
		
		// Создаем ответ об ошибке валидации в формате v2
		response := V2Response{
			Version: "v2",
			Content: V2Content{
				Type:     "instagram",
				Messages: []string{},
				Actions: []V2Action{
					{
						Action:    "set_field_value",
						FieldName: "error_message",
						Value:     "Неверные параметры запроса: " + err.Error(),
					},
					{
						Action:    "set_field_value",
						FieldName: "error_code",
						Value:     "invalid_request",
					},
					{
						Action:    "set_field_value",
						FieldName: "status",
						Value:     false,
					},
				},
			},
		}
		
		c.JSON(http.StatusOK, response)
		return
	}
	
	logger.WithFields(logrus.Fields{
		"link":   req.Link,
		"token":  req.Token,
		"trs":    req.TRS,
		"marker": req.Marker,
	}).Info("Обработка запроса getFromLink")
	
	// Делаем запрос к Travelpayouts API
	affiliateLink, tpError, err := makeAffiliateLink(req.Link, req.Token, req.TRS, req.Marker)
	if err != nil {
		logger.WithError(err).Error("Ошибка создания аффилиатной ссылки")
		
		// Создаем ответ об ошибке в формате v2
		response := V2Response{
			Version: "v2",
			Content: V2Content{
				Type:     "instagram",
				Messages: []string{},
				Actions: []V2Action{
					{
						Action:    "set_field_value",
						FieldName: "error_message",
						Value:     tpError.Message,
					},
					{
						Action:    "set_field_value",
						FieldName: "error_code",
						Value:     tpError.Code,
					},
					{
						Action:    "set_field_value",
						FieldName: "status",
						Value:     false,
					},
				},
			},
		}
		
		c.JSON(http.StatusOK, response)
		return
	}
	
	logger.WithField("affiliate_link", affiliateLink).Info("Аффилиатная ссылка создана успешно")
	
	// Создаем успешный ответ в формате v2
	response := V2Response{
		Version: "v2",
		Content: V2Content{
			Type:     "instagram",
			Messages: []string{},
			Actions: []V2Action{
				{
					Action:    "set_field_value",
					FieldName: "aff_url",
					Value:     affiliateLink,
				},
				{
					Action:    "set_field_value",
					FieldName: "status",
					Value:     true,
				},
			},
		},
	}
	
	c.JSON(http.StatusOK, response)
}

func getFromBrand(c *gin.Context) {
	var req GetFromBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Ошибка валидации запроса getFromBrand")
		
		// Создаем ответ об ошибке валидации в формате v2
		response := V2Response{
			Version: "v2",
			Content: V2Content{
				Type:     "instagram",
				Messages: []string{},
				Actions: []V2Action{
					{
						Action:    "set_field_value",
						FieldName: "error_message",
						Value:     "Неверные параметры запроса: " + err.Error(),
					},
					{
						Action:    "set_field_value",
						FieldName: "error_code",
						Value:     "invalid_request",
					},
					{
						Action:    "set_field_value",
						FieldName: "status",
						Value:     false,
					},
				},
			},
		}
		
		c.JSON(http.StatusOK, response)
		return
	}
	
	logger.WithFields(logrus.Fields{
		"brand_name": req.BrandName,
		"token":      req.Token,
		"trs":        req.TRS,
		"marker":     req.Marker,
	}).Info("Обработка запроса getFromBrand")
	
	// Поиск бренда в базе данных
	var brand Brand
	if err := db.Where("brand_name = ?", req.BrandName).First(&brand).Error; err != nil {
		logger.WithError(err).WithField("brand_name", req.BrandName).Error("Бренд не найден в базе данных")
		
		// Создаем ответ об ошибке в формате v2
		response := V2Response{
			Version: "v2",
			Content: V2Content{
				Type:     "instagram",
				Messages: []string{},
				Actions: []V2Action{
					{
						Action:    "set_field_value",
						FieldName: "error_message",
						Value:     "Бренд не найден: " + req.BrandName,
					},
					{
						Action:    "set_field_value",
						FieldName: "error_code",
						Value:     "brand_not_found",
					},
					{
						Action:    "set_field_value",
						FieldName: "status",
						Value:     false,
					},
				},
			},
		}
		
		c.JSON(http.StatusOK, response)
		return
	}
	
	logger.WithFields(logrus.Fields{
		"brand_name": brand.BrandName,
		"brand_link": brand.BrandLink,
	}).Info("Бренд найден в базе данных")
	
	// Делаем запрос к Travelpayouts API с ссылкой бренда
	affiliateLink, tpError, err := makeAffiliateLink(brand.BrandLink, req.Token, req.TRS, req.Marker)
	if err != nil {
		logger.WithError(err).Error("Ошибка создания аффилиатной ссылки для бренда")
		
		// Создаем ответ об ошибке в формате v2
		response := V2Response{
			Version: "v2",
			Content: V2Content{
				Type:     "instagram",
				Messages: []string{},
				Actions: []V2Action{
					{
						Action:    "set_field_value",
						FieldName: "error_message",
						Value:     tpError.Message,
					},
					{
						Action:    "set_field_value",
						FieldName: "error_code",
						Value:     tpError.Code,
					},
					{
						Action:    "set_field_value",
						FieldName: "status",
						Value:     false,
					},
				},
			},
		}
		
		c.JSON(http.StatusOK, response)
		return
	}
	
	logger.WithField("affiliate_link", affiliateLink).Info("Аффилиатная ссылка для бренда создана успешно")
	
	// Создаем успешный ответ в формате v2
	response := V2Response{
		Version: "v2",
		Content: V2Content{
			Type:     "instagram",
			Messages: []string{},
			Actions: []V2Action{
				{
					Action:    "set_field_value",
					FieldName: "aff_url",
					Value:     affiliateLink,
				},
				{
					Action:    "set_field_value",
					FieldName: "status",
					Value:     true,
				},
			},
		},
	}
	
	c.JSON(http.StatusOK, response)
} 