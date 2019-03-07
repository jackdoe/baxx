package config

type Config struct {
	MaxTokens    uint64
	DefaultQuota uint64
	SendGridKey  string
	Testing      bool
}

type StoreConfig struct {
	Endpoint        string
	Region          string
	Bucket          string
	TemporaryRoot   string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

var CONFIG = &Config{
	MaxTokens:    100,
	DefaultQuota: 10 * 1024 * 1024 * 1024,
	Testing:      false,
	SendGridKey:  "",
}
