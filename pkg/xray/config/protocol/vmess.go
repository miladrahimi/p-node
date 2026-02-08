package protocol

type VmessClient struct {
	Id       string `json:"id" validate:"required,min=1,max=64"`
	Email    string `json:"email,omitempty" validate:"omitempty,max=255"`
	Security string `json:"security,omitempty" validate:"omitempty,oneof=auto none zero aes-128-gcm chacha20-poly1305"`
}

type VmessOutboundServer struct {
	Address string         `json:"address" validate:"required"`
	Port    int            `json:"port" validate:"required,min=1,max=65535"`
	Users   []*VmessClient `json:"users,omitempty" validate:"omitempty,dive"`
}

type VmessOutboundSettings struct {
	Vnext []*VmessOutboundServer `json:"vnext,omitempty" validate:"omitempty,dive"`
}
