package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
)

func StartSession(c *gin.Context) {
	utils.JwtTokenParse(c)
}
