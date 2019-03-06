package main

import (
	"fmt"
	"github.com/badoux/checkmail"
)

func ValidateEmail(email string) error {
	err := checkmail.ValidateFormat(email)
	if err != nil {
		return fmt.Errorf("invalid email address (%s)", err.Error())
	}
	return nil
}

func ValidatePassword(p string) error {
	if len(p) < 8 {
		return fmt.Errorf("password is too short, refer to https://www.xkcd.com/936/")
	}
	return nil
}
