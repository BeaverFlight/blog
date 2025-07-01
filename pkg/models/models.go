package models

type Article struct {
	ID     int    `json:"id"`
	Author string `json:"author"`
	UserID int    `json:"user_id"`
	Text   string `json:"text"`
}

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Request struct {
	Article Article `json:"article"`
	User    User    `json:"user"`
}
