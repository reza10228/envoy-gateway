package jwt

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"main/build"
)
type JwtCustomClaims struct {
	UserId    uint64 `json:"user_id"`
	UID       string `json:"uid"`
	ParentId  uint64 `json:"parent_id"`
	RoleId    uint64 `json:"role_id"`
	PartnerId uint64 `json:"partner_id"`

	jwt.StandardClaims
}
func VerifyToken(Token string)  (*jwt.Token,JwtCustomClaims , error) {

	var claims JwtCustomClaims
	token , err := jwt.ParseWithClaims(Token,&claims, func(token *jwt.Token) (interface{}, error) {
		if _ , ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(build.Config.JwtAccessSecret) , nil
	})
	if err != nil {
		return nil,claims, err
	}


	return token,claims, nil
}

func TokenValidate(Token string) (JwtCustomClaims,error)  {
	token ,claims, err := VerifyToken(Token)
	if err != nil {
		return claims,err
	}
	if _ , ok := token.Claims.(jwt.Claims); !ok && !token.Valid{
		return claims,err
	}
	return  claims,nil
}
