package jwt

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"tigaputera-backend/sdk/error"
)

type jwtLib struct{}

type Interface interface {
	GetToken(interface{}) (string, error)
	DecodeToken(string) (map[string]interface{}, error)
}

func Init() Interface {
	return &jwtLib{}
}

func (j *jwtLib) GetToken(data interface{}) (string, error) {
	expTime, err := strconv.ParseInt(os.Getenv("JWT_EXPIRED_TIME_SEC"), 10, 64)
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data": data,
		"exp":  expTime + time.Now().Unix(),
	})

	return token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
}

func (j *jwtLib) DecodeToken(token string) (map[string]interface{}, error) {
	decoded, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})
	if err != nil {
		return nil, errors.Unauthorized("Invalid token")
	}

	claims, ok := decoded.Claims.(jwt.MapClaims)

	if !ok {
		return nil, errors.InternalServerError("Failed to decode token")
	}
	if !decoded.Valid {
		return nil, errors.Unauthorized("Invalid token")
	}

	return claims, nil
}
