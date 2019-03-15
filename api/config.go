package main

// in case of running in local mode
type LocalToken struct {
	UUID          string // has to map to baxx
	Bucket        string // bucket can be whatever you want
	EncryptionKey string // key to encrypt the data with
}

type Local struct {
	AllowedTokens []*LocalToken
}

type Config struct {
	MaxTokens         uint64
	DefaultQuota      uint64
	DefaultInodeQuota uint64
	SendGridKey       string
	Local             *Local
	SlackWebHook      string
}

var CONFIG = &Config{
	MaxTokens:         5,
	DefaultQuota:      10 * 1024 * 1024 * 1024,
	DefaultInodeQuota: 1000,
	SendGridKey:       "",
	SlackWebHook:      "",
}
