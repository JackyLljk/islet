package jwt_auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

// token authentication based on jwt_auth

type MyClaims struct {
	UserID uint64 `json:"user_id"`
	jwt.StandardClaims
}

// salting
var mySecret = []byte("El Psy Congroo")

const TokenExpireDuration = time.Hour * 24 * 365

func keyFunc(_ *jwt.Token) (i any, err error) {
	return mySecret, nil
}

func GenToken(userID uint64) (aToken, rToken string, err error) {
	c := MyClaims{
		userID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(),
			Issuer:    "islet",
		},
	}

	aToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(mySecret)

	rToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Second * 30).Unix(),
		Issuer:    "islet",
	}).SignedString(mySecret)

	return
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (claims *MyClaims, err error) {
	// 解析token
	var token *jwt.Token
	// 需要手动初始化，申请一块内存
	claims = new(MyClaims)

	token, err = jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return
	}

	if !token.Valid {
		err = errors.New("invalid token")
	}
	return
}

// RefreshToken 刷新AccessToken
func RefreshToken(aToken, rToken string) (newAToken, newRToken string, err error) {
	// 解析refresh token
	// refresh token无效直接返回
	if _, err = jwt.Parse(rToken, keyFunc); err != nil {
		return
	}

	// 从旧access token中解析出claims数据
	var claims MyClaims
	_, err = jwt.ParseWithClaims(aToken, &claims, keyFunc)

	// 检查是否是token invalid error
	var v *jwt.ValidationError
	_ = errors.As(err, &v)

	// 当access token是过期错误 并且 refresh token没有过期时就创建一个新的access token
	if v.Errors == jwt.ValidationErrorExpired {
		return GenToken(claims.UserID)
	}
	return
}
