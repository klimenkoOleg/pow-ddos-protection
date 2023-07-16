package server

/*
type Config struct {
	TCPAddress string

	Difficulty     byte
	ProofTokenSize int
}

func NewConfig() *Config {
	c := new(Config)

	c.TCPAddress = "0.0.0.0:9000" // envOrDefault("LISTEN_ADDR", "0.0.0.0:9000")

	c.Difficulty = 23     //byte(envOrDefaultInt("DIFFICULTY", 23))
	c.ProofTokenSize = 64 //envOrDefaultInt("PROOF_TOKEN_SIZE", 64)

	return c
}*/

/*func envOrDefault(key, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return value
}

func envOrDefaultInt(key string, defaultValue int) int {
	value, _ := os.LookupEnv(key)
	if v, err := strconv.Atoi(value); err == nil {
		return v
	}
	return defaultValue
}*/
