package config

const DefaultListenAddress = "127.0.0.1:8080"

type Settings struct {
	ListenAddress string
}

// Load returns the bootstrap server settings.
func Load() Settings {
	return Settings{ListenAddress: DefaultListenAddress}
}
