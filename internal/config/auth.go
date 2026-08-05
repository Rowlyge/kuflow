package config

// AuthConfig содержит настройки авторизации.
type AuthConfig struct {

	// HTTP-заголовок,
	// в котором клиент передаёт API Key.
	APIKeyHeader string
}

func loadAuthConfig() AuthConfig {

	return AuthConfig{

		APIKeyHeader: getEnv(
			"AUTH_API_KEY_HEADER",
			"X-API-Key",
		),
	}
}
