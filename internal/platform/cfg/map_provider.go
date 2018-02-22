package cfg

// MapProvider provides a simple implementation of the Provider whereby it just
// returns a stored map.
type MapProvider struct {
	Map map[string]string
}

// Provide implements the Provider interface.
func (mp MapProvider) Provide() (map[string]string, error) {
	return mp.Map, nil
}
