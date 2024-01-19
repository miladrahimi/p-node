package database

type Settings struct {
	XrayApiPort int    `json:"xray_api_port" validate:"required,min=1,max=65536"`
	HttpPort    int    `json:"http_port" validate:"required,min=1,max=65536"`
	HttpToken   string `json:"http_token" validate:"required,min=8,max=128"`
}
