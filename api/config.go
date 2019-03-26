package main

type Config struct {
	MaxTokens         uint64
	DefaultQuota      uint64
	DefaultInodeQuota uint64
	Bucket            string
}

var CONFIG = &Config{
	MaxTokens:         5,
	DefaultQuota:      10 * 1024 * 1024 * 1024,
	DefaultInodeQuota: 1000,
	Bucket:            "baxx",
}
