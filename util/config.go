package util

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	ForeignKeyViolation = "foreign_key_violation"
	UniqueViolation     = "unique_violation"
	TokenHasExpired     = "token has expired"
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

func loadEnvVariable(environmentName string) (string, error) {
	envVariable, ok := os.LookupEnv(environmentName)
	if ok {
		return envVariable, nil
	}

	envVariableFile, ok := os.LookupEnv(environmentName + "_FILE")
	if !ok {
		return "", fmt.Errorf("%s", "no "+environmentName+" or "+environmentName+"_FILE env var set")
	}

	data, err := os.ReadFile(envVariableFile)
	if err != nil {
		return "", fmt.Errorf("failed to read from var file: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

func LoadVault() (config *Config, err error) {

	dbDriver, err := loadEnvVariable("DB_DRIVER")
	if err != nil {
		return nil, err
	}

	dbSource, err := loadEnvVariable("DB_SOURCE")
	if err != nil {
		return nil, err
	}

	serverAddress, err := loadEnvVariable("SERVER_ADDRESS")
	if err != nil {
		return nil, err
	}
	// redisAddress, err := loadEnvVariable("REDIS_ADDRESS")
	// if err != nil {
	// 	return nil, err
	// }
	userTokenSymmetricKey, err := loadEnvVariable("USER_TOKEN_SYMMETRIC_KEY")
	if err != nil {
		return nil, err
	}
	adminTokenSymmetricKey, err := loadEnvVariable("ADMIN_TOKEN_SYMMETRIC_KEY")
	if err != nil {
		return nil, err
	}
	emailSenderName, err := loadEnvVariable("EMAIL_SENDER_NAME")
	if err != nil {
		return nil, err
	}
	emailSenderAddress, err := loadEnvVariable("EMAIL_SENDER_ADDRESS")
	if err != nil {
		return nil, err
	}
	emailSenderPassword, err := loadEnvVariable("EMAIL_SENDER_PASSWORD")
	if err != nil {
		return nil, err
	}
	imageKitPrivateKey, err := loadEnvVariable("IMAGE_KIT_PRIVATE_KEY")
	if err != nil {
		return nil, err
	}
	imageKitPublicKey, err := loadEnvVariable("IMAGE_KIT_PUBLIC_KEY")
	if err != nil {
		return nil, err
	}
	imageKitUrlEndPoint, err := loadEnvVariable("IMAGE_KIT_URL_ENDPOINT")
	if err != nil {
		return nil, err
	}
	accessTokenDurationValue, err := loadEnvVariable("ACCESS_TOKEN_DURATION")
	if err != nil {
		return nil, err
	}
	refreshTokenDurationValue, err := loadEnvVariable("REFRESH_TOKEN_DURATION")
	if err != nil {
		return nil, err
	}

	accessTokenDurration, err := time.ParseDuration(accessTokenDurationValue)
	if err != nil {
		return nil, err
	}

	refreshTokenDuration, err := time.ParseDuration(refreshTokenDurationValue)
	if err != nil {
		return nil, err
	}

	return &Config{
		DBDriver:               dbDriver,
		DBSource:               dbSource,
		ServerAddress:          serverAddress,
		RedisAddress:           "",
		UserTokenSymmetricKey:  userTokenSymmetricKey,
		AdminTokenSymmetricKey: adminTokenSymmetricKey,
		EmailSenderName:        emailSenderName,
		EmailSenderAddress:     emailSenderAddress,
		EmailSenderPassword:    emailSenderPassword,
		AccessTokenDuration:    accessTokenDurration,
		RefreshTokenDuration:   refreshTokenDuration,
		ImageKitPrivateKey:     imageKitPrivateKey,
		ImageKitPublicKey:      imageKitPublicKey,
		ImageKitUrlEndPoint:    imageKitUrlEndPoint,
	}, nil
}
