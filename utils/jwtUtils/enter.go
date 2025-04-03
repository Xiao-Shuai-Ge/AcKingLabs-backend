package jwtUtils

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"strings"
	"tgwp/global"
	"tgwp/log/zlog"
	"time"
)

type MyClaims struct {
	Userid int64  `json:"userid"`
	Class  string `json:"class"`
	jwt.RegisteredClaims
}

var mySecret = []byte("island")

type TokenData struct {
	Userid int64
	Class  string
	Time   time.Duration
}

func GenToken(data TokenData) (string, error) {
	// 创建一个我们自己的声明
	claims := MyClaims{
		data.Userid,
		data.Class,
		jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(data.Time)), // 过期时间
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(mySecret)
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (*MyClaims, error) {
	// 解析token
	if tokenString == "" {
		return nil, errors.New("token为空")
	}
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySecret, nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "token is expired") {
			return nil, errors.New("token已过期")
		}
		if strings.Contains(err.Error(), "signature is invalid") {
			return nil, errors.New("token无效")
		}
		if strings.Contains(err.Error(), "token contains an invalid") {
			return nil, errors.New("token非法")
		}
		return nil, err
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// 用于验证令牌是否有效
func IdentifyToken(ctx context.Context, Token string) (TokenData, error) {
	//解析token
	claim, err := ParseToken(Token)
	var data TokenData
	if err != nil {
		zlog.CtxErrorf(ctx, "IdentifyToken err: %v", err)
		return TokenData{}, err
	}
	data.Userid = claim.Userid
	data.Class = claim.Class
	if claim.Class == global.AUTH_ENUMS_RTOKEN {
		data.Time = global.RTOKEN_EFFECTIVE_TIME - time.Duration(time.Now().Unix()-claim.RegisteredClaims.NotBefore.Unix())
	} else {
		data.Time = global.ATOKEN_EFFECTIVE_TIME
	}
	return data, nil
}

func FullToken(class string, userid int64) (data TokenData) {
	//后期这两个都由雪花算法生成
	data.Userid = userid
	if class == global.AUTH_ENUMS_ATOKEN {
		data.Time = global.ATOKEN_EFFECTIVE_TIME
		data.Class = global.AUTH_ENUMS_ATOKEN
	} else {
		data.Time = global.RTOKEN_EFFECTIVE_TIME
		data.Class = global.AUTH_ENUMS_RTOKEN
	}
	return
}
