package util

import (
	"os"
	"time"
)

// config stores all configuration of the application
// The values are read by viper from a config file or environment variable.
type Config struct {
	DBDriver               string
	DBSource               string
	ServerAddress          string
	RedisAddress           string
	UserTokenSymmetricKey  string
	AdminTokenSymmetricKey string
	EmailSenderName        string
	EmailSenderAddress     string
	EmailSenderPassword    string
	AccessTokenDuration    time.Duration
	RefreshTokenDuration   time.Duration
	ImageKitPrivateKey     string
	ImageKitPublicKey      string
	ImageKitUrlEndPoint    string
}

// LoadConfig reads configuration from file or environment variable.
// func LoadConfig(path string) (config Config, err error) {
// 	viper.AddConfigPath(path)
// 	viper.SetConfigName(".env")
// 	viper.SetConfigType("env")

// 	viper.AutomaticEnv()

// 	err = viper.ReadInConfig()
// 	if err != nil {
// 		return
// 	}

// 	err = viper.Unmarshal(&config)
// 	return
// }

func LoadVault() (config Config, err error) {
	accessTokenDurration, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_DURATION"))
	if err != nil {
		return Config{}, err
	}

	refreshTokenDuration, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_DURATION"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		DBDriver:               os.Getenv("DB_DRIVER"),
		DBSource:               os.Getenv("DB_SOURCE"),
		ServerAddress:          os.Getenv("SERVER_ADDRESS"),
		RedisAddress:           os.Getenv("REDIS_ADDRESS"),
		UserTokenSymmetricKey:  os.Getenv("USER_TOKEN_SYMMETRIC_KEY"),
		AdminTokenSymmetricKey: os.Getenv("ADMIN_TOKEN_SYMMETRIC_KEY"),
		EmailSenderName:        os.Getenv("EMAIL_SENDER_NAME"),
		EmailSenderAddress:     os.Getenv("EMAIL_SENDER_ADDRESS"),
		EmailSenderPassword:    os.Getenv("EMAIL_SENDER_PASSWORD"),
		AccessTokenDuration:    accessTokenDurration,
		RefreshTokenDuration:   refreshTokenDuration,
		ImageKitPrivateKey:     os.Getenv("IMAGE_KIT_PRIVATE_KEY"),
		ImageKitPublicKey:      os.Getenv("IMAGE_KIT_PUBLIC_KEY"),
		ImageKitUrlEndPoint:    os.Getenv("IMAGE_KIT_URL_ENDPOINT"),
	}, nil
}
