package data

// Data represents the database schema.
type Data struct {
	Settings *Settings `json:"settings"`
	Manager  *Manager  `json:"manager"`
}

// New creates a new data instance.
func New(settings *Settings, manager *Manager) *Data {
	return &Data{
		Settings: settings,
		Manager:  manager,
	}
}

// Default returns the data with default values.
func Default() *Data {
	return New(
		DefaultSettings(),
		nil,
	)
}
