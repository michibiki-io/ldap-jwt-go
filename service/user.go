package service

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/michibiki-io/ldap-jwt-go/model"
	"github.com/michibiki-io/ldap-jwt-go/utility"
	"github.com/michibiki-io/ldap-jwt-go/utility/ldapc"
)

var ldapClient *ldapc.Client = nil

var redisClient *redis.Client = nil

func init() {

	// LDAP_PROTOCOL
	protocol := ldapc.LDAP
	switch utility.GetEnv("LDAP_PROTOCOL", "LDAP") {
	case "LDAPS":
		protocol = ldapc.LDAPS
	case "START_TLS":
		protocol = ldapc.START_TLS
	default:
		protocol = ldapc.LDAP
	}

	// LDAP_HOST
	ldapHost := utility.GetEnv("LDAP_HOST", "localhost")

	// LDAP_PORT
	ldapPort := utility.GetIntEnv("LDAP_PORT", 389)

	// LDAP_SKIPVERIFY
	// Skip verification of server certificate
	skipVerify := utility.GetBoolEnv("LDAP_SKIPVERIFY", false)

	// LDAP_BIND_DN
	bindDn := utility.GetEnv("LDAP_BIND_DN", "cn=readonly,dc=example,dc=com")

	// LDAP_BIND_PASSWORD
	bindPassword := utility.GetEnv("LDAP_BIND_PASSWORD", "readonly")

	// LDAP_BASE_DN
	baseDn := utility.GetEnv("LDAP_BASE_DN", "dc=example,dc=com")

	ldapClient = &ldapc.Client{
		Protocol:  protocol,
		Host:      ldapHost,
		Port:      ldapPort,
		TLSConfig: &tls.Config{InsecureSkipVerify: skipVerify},
		Bind: ldapc.Bind{
			BindDN:       bindDn,
			BindPassword: bindPassword,
			BaseDN:       baseDn,
		},
	}

	//Initializing redis
	dsn := utility.GetEnv("REDIS_HOST", "localhost:6379")
	redisClient = redis.NewClient(&redis.Options{
		Addr: dsn, //redis port
	})
	_, err := redisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
}

func verifyAuth(verifyKey *rsa.PublicKey, tokenString string, storeType model.StoreType) (
	token model.Token, expire_in int64, user model.User, storedAuth model.StoredAuth, error error) {

	token = model.Token{}
	user = model.User{}
	storedAuth = model.StoredAuth{}
	error = nil

	jwtService := JwtService{}

	if token, expire_in, error = jwtService.VerifyToken(verifyKey, tokenString); error != nil {
		return
	} else {
		storedAuth = model.StoredAuth{}
		if jsonObj, err := redisClient.Get(token.Uuid).Result(); err != nil {
			error = utility.NewError(fmt.Sprintf("Token is invalid"), utility.Unauthorized)
			utility.Log.Debug("Stored Token is not found, UUID: %s", token.Uuid)
		} else if err := json.Unmarshal([]byte(jsonObj), &storedAuth); err != nil {
			error = utility.NewError(fmt.Sprintf("Token is invalid"), utility.UnprocessableEntity)
			utility.Log.Debug("system cannot unmarshal the stored token, UUID: %s", token.Uuid)
		} else if storedAuth.Type != storeType {
			error = utility.NewError(fmt.Sprintf("Token is invalid"), utility.Unauthorized)
			utility.Log.Debug("Stored Token Type is different, UUID: %s", token.Uuid)
		} else if user, error = getUser(storedAuth.UserId); error != nil {
			return
		} else {
			error = nil
		}
		return
	}

}

func getUser(userId string) (user model.User, error error) {
	user = model.User{}
	error = nil

	// LDAP_FILTER_USER
	filter := utility.GetEnv("LDAP_FILTER_USER", "(&(objectClass=posixAccount)(uid=%s))")

	if entries, err := ldapClient.Search(filter, userId); err != nil {
		error = err
		return
	} else if len(entries) < 1 {
		error = utility.NewError(fmt.Sprintf("LDAP Authenticate failed: %v", err), utility.Unauthorized)
		return
	} else if len(entries) > 1 {
		error = utility.NewError(fmt.Sprintf("LDAP Authenticate failed: %v", err), utility.InternalServerError)
		utility.Log.Debug("Same UserId members are found: %v", err)
		return
	} else {
		user.DN = entries[0].DN
		user.Id = userId

		// LDAP_FILTER_GROUP
		filter := utility.GetEnv("LDAP_FILTER_GROUP", "(&(objectClass=groupOfNames)(member=%s))")

		if entries, error = ldapClient.Search(filter, user.DN); error == nil {
			for _, group := range entries {
				user.Groups = append(user.Groups, group.DN)
			}
		}
		return
	}
}

type UserService struct{}

func (s *UserService) Authorize(auth *model.Auth) (user model.User, error error) {
	user = model.User{}
	error = nil

	if ldapClient == nil {
		error = utility.NewError(fmt.Sprintf("cannot access authentication server"), utility.InternalServerError)
		utility.Log.Debug("LDAP client is not initialized, please check your LDAP server.")
		return
	}

	if tmpUser, err := getUser(auth.Username); error != nil {
		error = err
	} else if error = ldapClient.DoBind(tmpUser.DN, auth.Password); error != nil {
		error = utility.NewError(fmt.Sprintf("LDAP Authenticate failed, userId: %s", user.Id), utility.Unauthorized)
	} else {
		user = tmpUser
		error = nil
	}

	return
}

func (s *UserService) CreateAuth(user *model.User) (tokenSet model.TokenSet, expire_in model.ExpireIn, error error) {

	// return value
	tokenSet = model.TokenSet{}
	error = nil

	jwtService := JwtService{}

	if user.DN == "" {
		error = utility.NewError(fmt.Sprintf("LDAP Authenticate is not completed"), utility.Unauthorized)
		return
	}

	if ldapClient == nil {
		error = utility.NewError(fmt.Sprintf("cannot access authentication server"), utility.InternalServerError)
		utility.Log.Debug("LDAP client is not initialized, please check your LDAP server.")
		return
	}

	expiration := utility.GetIntEnv("ACCESS_TOKEN_EXPIRE", 15)
	if tokenSet.AccessToken, error = jwtService.CreateToken(accessTokenPair.SignKey, expiration); error != nil {
		tokenSet = model.TokenSet{}
		return
	}

	expiration = utility.GetIntEnv("REFRESH_TOKEN_EXIPIRE", 60*24*7)
	if tokenSet.RefreshToken, error = jwtService.CreateToken(refreshTokenPair.SignKey, expiration); error != nil {
		tokenSet = model.TokenSet{}
		return
	}

	at := time.Unix(tokenSet.AccessToken.Expires, 0)
	rt := time.Unix(tokenSet.RefreshToken.Expires, 0)
	now := time.Now()

	if jsonObj, err := json.Marshal(model.StoredAuth{UserId: user.Id, Type: model.StoreTypeAccess, LinkedUuid: tokenSet.RefreshToken.Uuid}); err != nil {
		error = err
		return
	} else if error = redisClient.Set(tokenSet.AccessToken.Uuid, jsonObj, at.Sub(now)).Err(); error != nil {
		return
	}

	if jsonObj, err := json.Marshal(model.StoredAuth{UserId: user.Id, Type: model.StoreTypeRefresh, LinkedUuid: tokenSet.AccessToken.Uuid}); err != nil {
		error = err
		return
	} else if error = redisClient.Set(tokenSet.RefreshToken.Uuid, jsonObj, rt.Sub(now)).Err(); error != nil {
		return
	}

	expire_in.AccessToken = int64(at.Sub(now).Seconds())
	expire_in.RefreshToken = int64(rt.Sub(now).Seconds())
	return

}

func (s *UserService) VerifyAuth(accessToken string) (user model.User, expire_in int64, error error) {

	user = model.User{}
	error = nil

	if _, expire_in, user, _, error = verifyAuth(accessTokenPair.VerifyKey, accessToken, model.StoreTypeAccess); error != nil {
		return
	} else {
		error = nil
		return
	}
}

func (s *UserService) RefreshAuth(refreshToken string) (tokenSet model.TokenSet, expire_in int64, error error) {

	tokenSet = model.TokenSet{}
	error = nil
	expire_in_ := model.ExpireIn{}

	if stRefreshToken, _, userFromRedis, storedAuth, err := verifyAuth(refreshTokenPair.VerifyKey, refreshToken, model.StoreTypeRefresh); err != nil {
		error = err
		return
	} else if deleted, err := redisClient.Del(stRefreshToken.Uuid).Result(); err != nil || deleted == 0 {
		error = utility.NewError(err.Error(), utility.Unauthorized)
		return
	} else if userFromLdap, err := getUser(userFromRedis.Id); err != nil {
		error = utility.NewError(fmt.Sprintf("LDAP Authenticate failed, userId: %s", userFromRedis.Id), utility.Unauthorized)
		return
	} else if userFromRedis.DN != userFromLdap.DN {
		error = utility.NewError(fmt.Sprintf("LDAP Authenticate failed, userId: %s", userFromRedis.Id), utility.Unauthorized)
		return
	} else if tokenSet, expire_in_, err = s.CreateAuth(&userFromLdap); err != nil {
		error = utility.NewError(fmt.Sprintf("LDAP Authenticate failed, userId: %s", userFromRedis.Id), utility.InternalServerError)
		utility.Log.Debug("CreateAuth is failed.")
		return
	} else {
		if deleted, err = redisClient.Del(storedAuth.LinkedUuid).Result(); err != nil || deleted == 0 {
			utility.Log.Debug("Deleting Linked Auth at Redis is failed, UUID: %s", storedAuth.LinkedUuid)
		}
		expire_in = expire_in_.RefreshToken
		error = nil
		return
	}
}

func (s *UserService) DeleteAuth(accessToken string) (error error) {

	error = nil

	if stAccessToken, _, _, storedAuth, err := verifyAuth(accessTokenPair.VerifyKey, accessToken, model.StoreTypeAccess); err != nil {
		error = err
		return
	} else {
		if deleted, err := redisClient.Del(stAccessToken.Uuid).Result(); err != nil || deleted == 0 {
			utility.Log.Debug("Deleting Stored Auth at Redis is failed, UUID: %s", stAccessToken.Uuid)
			error = err
		} else {
			if deleted, err = redisClient.Del(storedAuth.LinkedUuid).Result(); err != nil || deleted == 0 {
				utility.Log.Debug("Deleting Linked Auth at Redis is failed, UUID: %s", storedAuth.LinkedUuid)
			}
			error = nil
		}
		return
	}
}
