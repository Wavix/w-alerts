package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func IpCheckMiddleware(allowedIPs []string) gin.HandlerFunc {
	ipSet := make(map[string]struct{})
	for _, ip := range allowedIPs {
		ipSet[ip] = struct{}{}
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		_, ok := ipSet[clientIP]
		if !ok {
			Logger.Warn().Extra("client_ip", clientIP).Msg("IP address not allowed")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "IP address not allowed"})
			return
		}

		c.Next()
	}
}
