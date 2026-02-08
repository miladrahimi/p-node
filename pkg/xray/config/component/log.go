package component

type Log struct {
	LogLevel string `json:"loglevel" validate:"required"`
	Access   string `json:"access,omitempty"`
	Error    string `json:"error,omitempty"`
}
