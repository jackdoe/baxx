package config

type Config struct {
	MaxTokens         uint64
	DefaultQuota      uint64
	DefaultInodeQuota uint64
	SendGridKey       string
	Testing           bool
}

type StoreConfig struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	DisableSSL      bool
}

var CONFIG = &Config{
	MaxTokens:         5,
	DefaultQuota:      10 * 1024 * 1024 * 1024,
	DefaultInodeQuota: 1000,
	Testing:           false,
	SendGridKey:       "",
}
