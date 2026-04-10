package auth

import (
   "fmt"
   "time"

   "github.com/golang-jwt/jwt/v5"

)

type Claims struct {
	UserID string `json:"user_id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(secret , userID , email string , expiryHours int) (string , error){
	now := time.Now()
	expiresAt := now.Add(time.Duration(expiryHours) * time.Hour)

	claims := &Claims{
		UserID : userID,
		Email : email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}


func ParseToken(secret, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}