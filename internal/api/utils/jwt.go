package utils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/mekdep/server/internal/models"
)

var jwtSecretKey = "SecretYouShouldHide"

func JwtTokenParse(c *gin.Context) {
	defer c.Next()

	var jwtToken string
	queryToken, existsQueryToken := c.GetQuery("token")
	if !existsQueryToken {
		jwtToken, _ = extractBearerToken(c.GetHeader("Authorization"))
	} else {
		jwtToken = queryToken
	}
	if jwtToken == "" {
		return
	}

	token, err := parseToken(jwtToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "bad jwt token",
		})
		return
	}

	claims, OK := token.Claims.(jwt.MapClaims)
	if !OK {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unable to parse claims",
		})
		return
	}
	if claims["user_id"] == nil {
		return
	}
	userId := claims["user_id"].(string)
	schoolId := claims["school_id"].(string)
	if claims["period_id"] == nil {
		claims["period_id"] = ""
	}
	periodId := claims["period_id"].(string)
	role := claims["role_code"].(string)
	if userId == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unable to parse user id",
		})
		return
	}
	PrepareSession(c, jwtToken, string(userId), string(role), string(schoolId), string(periodId))
	SessionActByToken(jwtToken, time.Now())
}

func GenerateToken(c *gin.Context, userModel models.User, role *models.Role, schoolId *string, periodId *string, deviceToken *string) (jwt.MapClaims, *string, error) {
	if schoolId == nil {
		schoolId = new(string)
	}
	if role == nil {
		role = new(models.Role)
	}
	claims := jwt.MapClaims{}
	claims["exp"] = time.Now().Add(time.Hour * 24 * 120).Unix()
	claims["user_id"] = userModel.ID
	claims["school_id"] = *schoolId
	if periodId == nil {
		periodId = new(string)
	}
	claims["period_id"] = *periodId
	claims["role_code"] = role
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return nil, nil, err
	}
	claims["token"] = tokenStr
	ses, err := AddSession(c, claims, userModel, deviceToken)
	return claims, &ses.ID, err
}

func parseToken(signedToken string) (*jwt.Token, error) {
	parsedToken, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	return parsedToken, nil
}

func extractBearerToken(header string) (string, error) {
	if header == "" {
		return "", errors.New("bad header value given")
	}

	jwtToken := strings.Split(header, " ")
	if len(jwtToken) != 2 {
		return "", errors.New("incorrectly formatted authorization header")
	}

	return jwtToken[1], nil
}
