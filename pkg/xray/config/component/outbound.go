package component

type Outbound struct {
	Protocol       string          `json:"protocol" validate:"required"`
	Tag            string          `json:"tag" validate:"required"`
	Settings       any             `json:"settings,omitempty"`
	StreamSettings *StreamSettings `json:"streamSettings,omitempty"`
}

type StreamSettings struct {
	Network     string       `json:"network" validate:"required"`
	Security    string       `json:"security,omitempty"`
	TlsSettings *TlsSettings `json:"tlsSettings,omitempty"`
}

type TlsSettings struct {
	Fingerprint string `json:"fingerprint,omitempty"`
}
