package utils

import (
	"encoding/json"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/wavix/go-lib/logger"
)

var Logger = logger.New("Alerts")

func init() {
	env := "production"
	level := "info"
	plain := false

	err := godotenv.Load(".env")
	if err == nil {
		env = os.Getenv("ENV")
	}

	if env == "development" {
		level = "debug"
		plain = true
	}

	Logger = logger.New("Alerts", logger.SetupOptions{Plain: plain, MaxWordSize: 6, Level: logger.LogLevel(level)})
	Logger.Info().Msg("Logger initialized with level: " + level)
}

func GinLogger() gin.HandlerFunc {
	loggerFuncWithFormatter := func(c *gin.Context, params gin.LogFormatterParams) string {
		var body string
		requestBody, exists := c.Get("RequestBody")

		if exists {
			json, err := convertRequestBodyToJSON(requestBody.([]byte))
			if err == nil {
				body = json
			}

		}

		logger := Logger.Context("API").
			Info().
			Extra("path", params.Path).
			Extra("method", params.Method).
			Extra("status", params.StatusCode).
			Extra("client_ip", params.ClientIP).
			Extra("latency", params.Latency).
			Extra("user_agent", params.Request.UserAgent()).
			Extra("body", body)

		logger.Info().Msg("Http request received")

		return ""
	}

	return func(c *gin.Context) {
		c.Next()

		gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
			return loggerFuncWithFormatter(c, params)
		})(c)
	}
}

func convertRequestBodyToJSON(requestBody []byte) (string, error) {
	var parsedData map[string]interface{}

	err := json.Unmarshal(requestBody, &parsedData)
	if err != nil {
		return "", err
	}

	jsonString, err := json.Marshal(parsedData)
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
