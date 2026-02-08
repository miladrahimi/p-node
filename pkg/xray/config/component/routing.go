package component

type Rule struct {
	InboundTag  []string `json:"inboundTag" validate:"required"`
	OutboundTag string   `json:"outboundTag,omitempty"`
	BalancerTag string   `json:"balancerTag,omitempty"`
	Domain      []string `json:"domain,omitempty"`
}

type RoutingSettings struct {
	Rules []*Rule `json:"rules" validate:"required,dive"`
}

type Balancer struct {
	Tag      string   `json:"tag" validate:"required"`
	Selector []string `json:"selector"`
}

type Routing struct {
	DomainStrategy string      `json:"domainStrategy" validate:"required"`
	DomainMatcher  string      `json:"domainMatcher" validate:"required"`
	Rules          []*Rule     `json:"rules,omitempty" validate:"omitempty,dive"`
	Balancers      []*Balancer `json:"balancers,omitempty" validate:"omitempty,dive"`
}

type Reverse struct {
	Bridges []*ReverseItem `json:"bridges,omitempty"  validate:"omitempty,dive"`
	Portals []*ReverseItem `json:"portals,omitempty"  validate:"omitempty,dive"`
}

type ReverseItem struct {
	Tag    string `json:"tag"  validate:"required"`
	Domain string `json:"domain"  validate:"required"`
}
