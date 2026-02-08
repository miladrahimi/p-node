package protocol

type SsClient struct {
	Password string `json:"password" validate:"required,min=1,max=64"`
	Method   string `json:"method" validate:"required"`
	Email    string `json:"email" validate:"required"`
}

type SsOutboundServer struct {
	Address  string `json:"address" validate:"required"`
	Port     int    `json:"port" validate:"required,min=1,max=65535"`
	Method   string `json:"method" validate:"required"`
	Password string `json:"password" validate:"required"`
	Uot      bool   `json:"uot"`
}

type SsOutboundSettings struct {
	Servers []*SsOutboundServer `json:"servers,omitempty" validate:"omitempty,dive"`
}
