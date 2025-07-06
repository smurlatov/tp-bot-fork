package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"tp-go-service/modules/ManyChat"
	"tp-go-service/modules/TravelPayouts"
	"tp-go-service/modules/WeGoTrip"
)

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

type GetFeedRequest struct {
	City     string `json:"city" binding:"required"`
	Lang     string `json:"lang"`
	Currency string `json:"currency"`
	Page     int    `json:"page"`
}

var logger *logrus.Logger

func main() {

	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.LoggerWithWriter(logger.Out), gin.Recovery())

	r.Use(func(c *gin.Context) {
		logger.WithFields(logrus.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
		}).Info("HTTP Request")
		c.Next()
	})

	api := r.Group("/api")
	{
		api.POST("/getFromLink", getFromLink)
		api.POST("/getFeed", getFeed)
	}

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

func getFromLink(c *gin.Context) {
	var req GetFromLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Ошибка валидации запроса getFromLink")

		mc := ManyChat.New()
		response := mc.FromValidationError("Неверные параметры запроса: " + err.Error())

		c.JSON(http.StatusOK, response)
		return
	}

	logger.WithFields(logrus.Fields{
		"link":   req.Link,
		"token":  req.Token,
		"trs":    req.TRS,
		"marker": req.Marker,
	}).Info("Обработка запроса getFromLink")

	tp, err := TravelPayouts.New(req.Token, req.TRS, req.Marker)
	if err != nil {
		logger.WithError(err).Error("Ошибка создания TravelPayouts клиента")

		mc := ManyChat.New()
		response := mc.FromError(err)

		c.JSON(http.StatusOK, response)
		return
	}

	affiliateLink, err := tp.GetFromLink(req.Link)
	if err != nil {
		logger.WithError(err).Error("Ошибка создания аффилиатной ссылки")

		mc := ManyChat.New()
		response := mc.FromError(err)

		c.JSON(http.StatusOK, response)
		return
	}

	logger.WithField("affiliate_link", affiliateLink).Info("Аффилиатная ссылка создана успешно")

	mc := ManyChat.New()
	response := mc.FromTravelPayoutsResponse(affiliateLink)

	c.JSON(http.StatusOK, response)
}

func getFeed(c *gin.Context) {
	var req GetFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Ошибка валидации запроса getFeed")

		mc := ManyChat.New()
		response := mc.FromValidationError("Неверные параметры запроса: " + err.Error())

		c.JSON(http.StatusOK, response)
		return
	}

	logger.WithFields(logrus.Fields{
		"city":     req.City,
		"lang":     req.Lang,
		"currency": req.Currency,
		"page":     req.Page,
	}).Info("Обработка запроса getFeed")

	wg := WeGoTrip.New()

	feed, err := wg.GetFeed(req.City, req.Lang, req.Currency, req.Page)
	if err != nil {
		logger.WithError(err).Error("Ошибка получения данных о поездках")

		mc := ManyChat.New()
		response := mc.FromError(err)

		c.JSON(http.StatusOK, response)
		return
	}

	logger.WithField("feed_length", len(feed)).Info("Данные о поездках получены успешно")

	mc := ManyChat.New()
	response := mc.FromWeGoGetRespose(feed)

	c.JSON(http.StatusOK, response)
}
