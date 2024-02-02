package database

type Settings struct {
	HttpPort  int    `json:"http_port" validate:"required,min=1,max=65536"`
	HttpToken string `json:"http_token" validate:"required,min=8,max=128"`
}
