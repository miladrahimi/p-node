package component

// Paqet is the paqet (https://github.com/hanselime/paqet) server a P-Node runs
// next to Xray for Relay RR2PQ. It travels inside the Xray config so pushing
// and pulling deliver it along with everything else; Xray ignores the key.
type Paqet struct {
	Port int    `json:"port" validate:"required,min=1,max=65535"`
	Key  string `json:"key" validate:"required"`
}
