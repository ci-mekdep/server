package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/utils"
	"github.com/sirupsen/logrus"
)

func SetLoggerRequest(c *gin.Context) {
	c.Request.ParseForm()
	utils.Logger = utils.Logger.WithFields(logrus.Fields{
		"url":   c.Request.Method + " " + c.Request.RequestURI,
		"form":  c.Request.PostForm,
		"ip":    c.ClientIP(),
		"agent": c.Request.UserAgent(),
	})
}
