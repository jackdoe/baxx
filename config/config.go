package config

type Config struct {
	FileRoot     string
	MaxTokens    uint64
	DefaultQuota uint64
	Testing      bool
}

var CONFIG = &Config{
	FileRoot:     "/tmp",
	MaxTokens:    100,
	DefaultQuota: 10 * 1024 * 1024 * 1024,
	Testing:      false,
}
