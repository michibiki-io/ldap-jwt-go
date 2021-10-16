package model

type Token struct {
	Token   string
	Uuid    string
	Expires int64
}

type TokenSet struct {
	AccessToken  Token
	RefreshToken Token
}

type ExpireIn struct {
	AccessToken  int64
	RefreshToken int64
}
