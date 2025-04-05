package jwtUtils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"tgwp/global"
	"time"
)

type TokenData struct {
	Userid   string `json:"user_id"`
	Username string `json:"username"`
	Class    string `json:"class"`
}

func GenToken(userid string, username string, exp time.Duration, class string) (string, error) {
	claims := jwt.MapClaims{
		"userid":   userid,
		"username": username,
		"class":    class,
		"exp":      time.Now().Add(exp).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(global.Config.JWT.Secret))
	return tokenString, err
}

func GenAtoken(userid string, username string, exp time.Duration) (string, error) {
	return GenToken(userid, username, exp, global.AUTH_ENUMS_ATOKEN)
}

func GenRtoken(userid string, username string, exp time.Duration) (string, error) {
	return GenToken(userid, username, exp, global.AUTH_ENUMS_RTOKEN)
}

func IdentifyToken(tokenString string) (TokenData, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名方法: %v", token.Header["alg"])
		}
		return []byte(global.Config.JWT.Secret), nil
	})
	if err != nil {
		return TokenData{}, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		// 验证token是否过期
		if time.Now().Unix() > int64(claims["exp"].(float64)) {
			return TokenData{}, fmt.Errorf("token已过期")
		}
	} else {
		// 解析失败
		return TokenData{}, fmt.Errorf("无效的token")
	}
	// 解析token成功
	return TokenData{
		Userid:   claims["userid"].(string),
		Username: claims["username"].(string),
		Class:    claims["class"].(string),
	}, nil
}

func GetUserId(c *gin.Context) string {
	if data, exists := c.Get(global.TOKEN_USER_ID); exists {
		userId, ok := data.(string)
		if ok {
			return userId
		}
	}
	return ""
}
