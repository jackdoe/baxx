package common

import uuid "github.com/satori/go.uuid"

func GetUUID() string {
	return uuid.Must(uuid.NewV4()).String()
}
