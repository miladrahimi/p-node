package component

type Policy struct {
	Levels map[string]PolicyLevel `json:"levels"`
	System PolicySystem           `json:"system"`
}

type PolicyLevel struct {
	StatsUserUplink   bool `json:"statsUserUplink"`
	StatsUserDownlink bool `json:"statsUserDownlink"`
}

type PolicySystem struct {
	StatsInboundUplink    bool `json:"statsInboundUplink"`
	StatsInboundDownlink  bool `json:"statsInboundDownlink"`
	StatsOutboundUplink   bool `json:"statsOutboundUplink"`
	StatsOutboundDownlink bool `json:"statsOutboundDownlink"`
}
