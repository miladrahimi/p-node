package data

// Manager is the manager struct.
type Manager struct {
	Url   string `json:"url" validate:"required,url,min=1,max=1024"` // P-Manager URL
	Token string `json:"token" validate:"required,min=1,max=128"`    // P-Manager authentication token
}

// NewManager returns the manager with the given url and token.
func NewManager(url string, token string) *Manager {
	return &Manager{
		Url:   url,
		Token: token,
	}
}
