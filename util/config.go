package util

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

// config stores all configuration of the application
// The values are read by viper from a config file or environment variable.
type Config struct {
	DBDriver               string        `mapstructure:"DB_DRIVER"`
	DBSource               string        `mapstructure:"DB_SOURCE"`
	ServerAddress          string        `mapstructure:"SERVER_ADDRESS"`
	RedisAddress           string        `mapstructure:"REDIS_ADDRESS"`
	UserTokenSymmetricKey  string        `mapstructure:"USER_TOKEN_SYMMETRIC_KEY"`
	AdminTokenSymmetricKey string        `mapstructure:"ADMIN_TOKEN_SYMMETRIC_KEY"`
	EmailSenderName        string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress     string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword    string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
	AccessTokenDuration    time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration   time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	ImageKitPrivateKey     string        `mapstructure:"IMAGE_KIT_PRIVATE_KEY"`
	ImageKitPublicKey      string        `mapstructure:"IMAGE_KIT_PUBLIC_KEY"`
	ImageKitUrlEndPoint    string        `mapstructure:"IMAGE_KIT_URL_ENDPOINT"`
}

// LoadConfig reads configuration from file or environment variable.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

func LoadVault() (config Config, err error) {
	// err = godotenv.Load(filename...)
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }
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
