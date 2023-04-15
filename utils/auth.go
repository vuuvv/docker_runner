package utils

import "vuuvv.cn/unisoftcn/orca/auth"

/*
{
"jwtIssuer": "orca.vuuvv.com",
"jwtSecret": "eyJhbG.JIUzI1NiIsInR5cCI6IkpXVCJ9",
"jwtTokenPrefix": "Bearer",
"accessTokenMaxAge": 15,
"accessTokenHead": "Authorization",
"refreshTokenMaxAge": 60,
"refreshTokenHead": "RefreshToken"
}
*/

func GetAuthConfig() *auth.Config {
	return &auth.Config{
		// 用于生成token的密钥
		JwtIssuer:          "orca.vuuvv.com",
		JwtSecret:          "eyJhbG.JIUzI1NiIsInR5cCI6IkpXVCJ9",
		JwtTokenPrefix:     "Bearer",
		AccessTokenMaxAge:  15,
		AccessTokenHead:    "Authorization",
		RefreshTokenMaxAge: 60,
		RefreshTokenHead:   "RefreshToken",
	}
}
