package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaim struct {
	UserID 		uuid.UUID		`json:"user_id"`
	Username 	string			`json:"username"`
	Role 		string			`json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

//Generate AccessToken - Tạo Access Token
func GenerateAccessToken(userID uuid.UUID, username,role, secret string, expiresMinute int) (string,error){
	claims := JWTClaim{
		UserID: userID,
		Username: username,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(expiresMinute))),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

//Generate RefreshToken - Tạo Refresh Token
func GenerateRefreshToken(userID uuid.UUID, secret string, expiresDays int) (string,error){
	claims := jwt.RegisteredClaims{
		Subject: userID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * time.Duration(expiresDays))),
		IssuedAt: jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}


//Verify AccessToken - Kiểm tra Access Token
func VerifyAccessToken(tokenString, secret string) (*JWTClaim, error){
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaim{}, func(token *jwt.Token) (interface{}, error){
		return []byte(secret), nil
	})
	if err !=nil{
		return nil,err
	}

	if claims, ok := token.Claims.(*JWTClaim); ok && token.Valid{
		return claims, nil
	}

	return nil,errors.New("Token Không Hợp Lệ")
}

//Verify RefreshToken - Kiểm tra Refresh Token
func VerifyRefreshToken(tokenString, secret string) (uuid.UUID, error){
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error){
		return []byte(secret), nil
	})
	if err !=nil{
		return uuid.Nil,err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid{
		return uuid.Parse(claims.Subject)
	}

	return uuid.Nil,errors.New("Token Không Hợp Lệ")
}
