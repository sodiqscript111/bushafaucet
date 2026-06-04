package config

import "os"

type Config struct {
	ServerPort	string

	RedisAddr	string

	DatabaseURL	string

	BushaAPIKey	string
	BushaBaseURL	string
	BushaProfileID	string

	MaxFaucetAmountBTC	string
	MaxFaucetAmountETH	string
	MaxFaucetAmountUSDT	string
	MaxFaucetAmountUSDC	string
	MaxFaucetAmountBNB	string
}

func SupportedBlockchains() []string {
	return []string{"BTC", "ETH", "USDT", "USDC", "BNB"}
}

var BlockchainNetworks = map[string]string{
	"BTC":	"BTC",
	"ETH":	"ETH",
	"USDT":	"ETH",
	"USDC":	"ETH",
	"BNB":	"BNB",
}

func Load() *Config {
	return &Config{
		ServerPort:	getEnv("SERVER_PORT", "8085"),

		RedisAddr:	getEnv("REDIS_ADDR", "127.0.0.1:6380"),
		DatabaseURL:	getEnv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5433/faucet?sslmode=disable"),

		BushaAPIKey:	getEnv("BUSHA_API_KEY", ""),
		BushaBaseURL:	getEnv("BUSHA_BASE_URL", "https://api.staging.busha.so"),
		BushaProfileID:	getEnv("BUSHA_PROFILE_ID", ""),

		MaxFaucetAmountBTC:	getEnv("FAUCET_AMOUNT_BTC", "0.0001"),
		MaxFaucetAmountETH:	getEnv("FAUCET_AMOUNT_ETH", "0.005"),
		MaxFaucetAmountUSDT:	getEnv("FAUCET_AMOUNT_USDT", "5"),
		MaxFaucetAmountUSDC:	getEnv("FAUCET_AMOUNT_USDC", "5"),
		MaxFaucetAmountBNB:	getEnv("FAUCET_AMOUNT_BNB", "0.01"),
	}
}

func (c *Config) MaxFaucetAmount(blockchain string) string {
	switch blockchain {
	case "BTC":
		return c.MaxFaucetAmountBTC
	case "ETH":
		return c.MaxFaucetAmountETH
	case "USDT":
		return c.MaxFaucetAmountUSDT
	case "USDC":
		return c.MaxFaucetAmountUSDC
	case "BNB":
		return c.MaxFaucetAmountBNB
	default:
		return "0"
	}
}

func IsSupportedBlockchain(bc string) bool {
	for _, s := range SupportedBlockchains() {
		if s == bc {
			return true
		}
	}
	return false
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
