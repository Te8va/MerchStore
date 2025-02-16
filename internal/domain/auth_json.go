package domain

type User struct {
	Username string
	Password string
	Token    string
}

type AuthorizationData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}
