package database

type Manager struct {
	Url   string `json:"url" validate:"required,url,min=1,max=1024"`
	Token string `json:"token" validate:"required,min=1,max=128"`
}
