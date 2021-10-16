package service

import (
	"crypto/rsa"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/michibiki-io/ldap-jwt-go/model"
	"github.com/michibiki-io/ldap-jwt-go/utility"
	"github.com/twinj/uuid"
)

var (
	accessTokenPair  *model.Rsa
	refreshTokenPair *model.Rsa
)

func init() {
	if accessTokenPair == nil {
		accessTokenPair = &model.Rsa{}
		accessTokenPair.Load("private/access.key", "private/access.key.pub")
	}
	if refreshTokenPair == nil {
		refreshTokenPair = &model.Rsa{}
		refreshTokenPair.Load("private/refresh.key", "private/refresh.key.pub")
	}
}

type JwtService struct{}

func (*JwtService) CreateToken(signKey *rsa.PrivateKey, expiration int) (stToken model.Token, createError error) {

	stToken = model.Token{}
	createError = nil

	token := jwt.New(jwt.SigningMethodRS512)

	// set claims
	claims := token.Claims.(jwt.MapClaims)
	// n minutes
	expired := time.Now().UTC().Add(time.Minute * time.Duration(expiration)).Unix()
	//expired := time.Now().UTC().Add(time.Second * time.Duration(expiration)).Unix()
	claims["exp"] = expired
	claims["uuid"] = uuid.NewV4().String()

	if stToken.Token, createError = token.SignedString(signKey); createError != nil {
		return
	} else {
		createError = nil
		stToken.Expires = expired
		stToken.Uuid = claims["uuid"].(string)
		return
	}
}

func (*JwtService) VerifyToken(verifyKey *rsa.PublicKey, tokenString string) (stToken model.Token, expire_in int64, verifyError error) {

	stToken = model.Token{}
	verifyError = nil

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, utility.NewError(fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]), utility.Forbidden)
		} else {
			return verifyKey, nil
		}
	})

	if token.Valid {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if expire, ok := claims["exp"].(float64); !ok {
				verifyError = utility.NewError(fmt.Sprintf("Unexpected exp data type."), utility.UnprocessableEntity)
			} else {
				stToken.Expires = int64(expire)
				expire := time.Unix(stToken.Expires, 0)
				now := time.Now()
				expire_in = int64(expire.Sub(now).Seconds())
			}
			if stToken.Uuid, ok = claims["uuid"].(string); !ok {
				verifyError = utility.NewError(fmt.Sprintf("Unexpected uuid data type."), utility.UnprocessableEntity)
			}
			return
		}
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			verifyError = utility.NewError(fmt.Sprintf("Token is not jwt token"), utility.UnprocessableEntity)
			return
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			// Token is either expired or not active yet
			verifyError = utility.NewError(fmt.Sprintf("Token is expired"), utility.Expired)
		} else {
			verifyError = utility.NewError(fmt.Sprintf("Token is not valid"), utility.Unauthorized)
		}
	} else {
		verifyError = utility.NewError(fmt.Sprintf("Token is not valid"), utility.Unauthorized)
	}

	return
}
