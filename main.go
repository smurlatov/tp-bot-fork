package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
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
		c.JSON(http.StatusBadRequest, Response{Error: "Неверные параметры запроса: " + err.Error()})
		return
	}
	
	logger.WithFields(logrus.Fields{
		"link":   req.Link,
		"token":  req.Token,
		"trs":    req.TRS,
		"marker": req.Marker,
	}).Info("Обработка запроса getFromLink")
	
	// Делаем запрос к Travelpayouts API
	affiliateLink, err := makeAffiliateLink(req.Link, req.Token, req.TRS, req.Marker)
	if err != nil {
		logger.WithError(err).Error("Ошибка создания аффилиатной ссылки")
		c.JSON(http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	
	logger.WithField("affiliate_link", affiliateLink).Info("Аффилиатная ссылка создана успешно")
	c.JSON(http.StatusOK, Response{Link: affiliateLink})
}

func getFromBrand(c *gin.Context) {
	var req GetFromBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Ошибка валидации запроса getFromBrand")
		c.JSON(http.StatusBadRequest, Response{Error: "Неверные параметры запроса: " + err.Error()})
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
		c.JSON(http.StatusNotFound, Response{Error: "Бренд не найден: " + req.BrandName})
		return
	}
	
	logger.WithFields(logrus.Fields{
		"brand_name": brand.BrandName,
		"brand_link": brand.BrandLink,
	}).Info("Бренд найден в базе данных")
	
	// Делаем запрос к Travelpayouts API с ссылкой бренда
	affiliateLink, err := makeAffiliateLink(brand.BrandLink, req.Token, req.TRS, req.Marker)
	if err != nil {
		logger.WithError(err).Error("Ошибка создания аффилиатной ссылки для бренда")
		c.JSON(http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	
	logger.WithField("affiliate_link", affiliateLink).Info("Аффилиатная ссылка для бренда создана успешно")
	c.JSON(http.StatusOK, Response{Link: affiliateLink})
} 