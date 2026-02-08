package component

type Inbound struct {
	Listen         string           `json:"listen,omitempty"`
	Port           int              `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
	Protocol       string           `json:"protocol" validate:"required,oneof=vless dokodemo-door socks"`
	Settings       *InboundSettings `json:"settings" validate:"required"`
	StreamSettings *StreamSettings  `json:"streamSettings,omitempty"`
	Sniffing       *Sniffing        `json:"sniffing,omitempty"`
	Tag            string           `json:"tag,omitempty"`
}

type InboundSettings struct {
	Clients    []*VlessUser     `json:"clients,omitempty" validate:"omitempty,dive"`
	Decryption string           `json:"decryption,omitempty" validate:"omitempty,oneof=none"`
	Fallbacks  []*VlessFallback `json:"fallbacks,omitempty" validate:"omitempty,dive"`

	Address string `json:"address,omitempty"`
	Port    int    `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
	Network string `json:"network,omitempty"`
	Udp     bool   `json:"udp,omitempty"`
}

type Outbound struct {
	Protocol       string            `json:"protocol" validate:"required,oneof=freedom blackhole vless"`
	Tag            string            `json:"tag,omitempty"`
	Settings       *OutboundSettings `json:"settings,omitempty" validate:"omitempty"`
	StreamSettings *StreamSettings   `json:"streamSettings,omitempty"`
}

type StreamSettings struct {
	Network         string           `json:"network" validate:"required,oneof=tcp ws grpc raw xhttp"`
	Security        string           `json:"security,omitempty" validate:"omitempty,oneof=tls reality"`
	TlsSettings     *TlsSettings     `json:"tlsSettings,omitempty"`
	RealitySettings *RealitySettings `json:"realitySettings,omitempty"`
	WsSettings      *WsSettings      `json:"wsSettings,omitempty"`
	GrpcSettings    *GrpcSettings    `json:"grpcSettings,omitempty"`
	XhttpSettings   *XhttpSettings   `json:"xhttpSettings,omitempty"`
}

type TlsSettings struct {
	ServerName   string         `json:"serverName,omitempty"`
	Alpn         []string       `json:"alpn,omitempty" validate:"omitempty,dive,min=1"`
	Certificates []*Certificate `json:"certificates,omitempty" validate:"omitempty,dive"`
	Fingerprint  string         `json:"fingerprint,omitempty"`
}

type Certificate struct {
	CertificateFile string `json:"certificateFile" validate:"required"`
	KeyFile         string `json:"keyFile" validate:"required"`
}

type RealitySettings struct {
	Dest        string   `json:"dest,omitempty"`
	ServerNames []string `json:"serverNames,omitempty"`
	PrivateKey  string   `json:"privateKey,omitempty"`
	ShortIds    []string `json:"shortIds,omitempty"`

	Fingerprint string `json:"fingerprint,omitempty"`
	ServerName  string `json:"serverName,omitempty"`
	PublicKey   string `json:"publicKey,omitempty"`
	Password    string `json:"password,omitempty"`
	SpiderX     string `json:"spiderX,omitempty"`
	ShortId     string `json:"shortId,omitempty"`

	Show            bool       `json:"show,omitempty"`
	Target          string     `json:"target,omitempty"`
	Xver            int        `json:"xver,omitempty" validate:"omitempty,min=0"`
	MinClientVer    string     `json:"minClientVer,omitempty"`
	MaxClientVer    string     `json:"maxClientVer,omitempty"`
	MaxTimeDiff     int        `json:"maxTimeDiff,omitempty" validate:"omitempty,min=0"`
	Mldsa65Seed     string     `json:"mldsa65Seed,omitempty"`
	Mldsa65Verify   string     `json:"mldsa65Verify,omitempty"`
	LimitFallbackUp *RateLimit `json:"limitFallbackUpload,omitempty"`
	LimitFallbackDn *RateLimit `json:"limitFallbackDownload,omitempty"`
}

type Sniffing struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride,omitempty"`
	RouteOnly    bool     `json:"routeOnly,omitempty"`
}

type WsSettings struct {
	Path string `json:"path" validate:"required"`
}

type GrpcSettings struct {
	ServiceName string `json:"serviceName" validate:"required"`
}

type XhttpSettings struct {
	Path string `json:"path" validate:"required"`
}

type VlessUser struct {
	Id         string `json:"id" validate:"required,min=1,max=64"`
	Flow       string `json:"flow,omitempty" validate:"omitempty,max=64"`
	Email      string `json:"email,omitempty" validate:"omitempty,max=255"`
	Level      int    `json:"level,omitempty" validate:"omitempty,min=0"`
	Encryption string `json:"encryption,omitempty" validate:"omitempty,oneof=none"`
}

type VlessOutboundServer struct {
	Address string       `json:"address" validate:"required"`
	Port    int          `json:"port" validate:"required,min=1,max=65535"`
	Users   []*VlessUser `json:"users,omitempty" validate:"omitempty,dive"`
}

type OutboundSettings struct {
	Vnext []*VlessOutboundServer `json:"vnext,omitempty" validate:"omitempty,dive"`
}

type VlessFallback struct {
	Dest int    `json:"dest,omitempty" validate:"omitempty,min=1,max=65535"`
	Alpn string `json:"alpn,omitempty"`
}

type RateLimit struct {
	AfterBytes       int `json:"afterBytes,omitempty" validate:"omitempty,min=0"`
	BytesPerSec      int `json:"bytesPerSec,omitempty" validate:"omitempty,min=0"`
	BurstBytesPerSec int `json:"burstBytesPerSec,omitempty" validate:"omitempty,min=0"`
}
