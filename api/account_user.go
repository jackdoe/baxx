package main

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

func hashAndSalt(pwd string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	return string(hash)
}

func comparePasswords(hashedPwd string, plainPwd string) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(plainPwd))
	return err == nil
}

type PaymentHistory struct {
	ID        uint64    `gorm:"primary_key" json:"-"`
	UserID    uint64    `gorm:"type:bigint not null REFERENCES users(id)" json:"-"`
	IPN       string    `gorm:"not null;type:text" json:"-"`
	IPNRAW    string    `gorm:"not null;type:text" json:"-"`
	UpdatedAt time.Time `json:"-"`
	CreatedAt time.Time `json:"-"`
}

type VerificationLink struct {
	ID         string     `gorm:"primary_key" json:"-"`
	UserID     uint64     `gorm:"type:bigint not null REFERENCES users(id)" json:"-"`
	Email      string     `gorm:"not null" json:"-"`
	VerifiedAt *time.Time `gorm:"null" json:"-"`
	SentAt     uint64     `gorm:"not null" json:"-"`
	UpdatedAt  time.Time  `json:"-"`
	CreatedAt  time.Time  `json:"-"`
}

type User struct {
	ID                    uint64     `gorm:"primary_key" json:"-"`
	PaymentID             string     `gorm:"not null" json:"-"`
	Email                 string     `gorm:"not null" json:"-"`
	EmailVerified         *time.Time `json:"-"`
	StartedSubscription   *time.Time `json:"-"`
	CancelledSubscription *time.Time `json:"-"`
	HashedPassword        string     `gorm:"not null" json:"-"`
	CreatedAt             time.Time  `json:"-"`
	UpdatedAt             time.Time  `json:"-"`
}

func (user *User) Paid() bool {
	if user.StartedSubscription == nil {
		return false
	}
	if user.CancelledSubscription == nil {
		return true
	}

	delta := user.CancelledSubscription.Sub(*user.StartedSubscription)
	return delta.Hours() < (24 * 30)

}

func (user *User) GenerateVerificationLink() *VerificationLink {
	return &VerificationLink{
		ID:     getUUID(),
		UserID: user.ID,
		Email:  user.Email,
	}
}

func (user *User) MatchPassword(p string) bool {
	return comparePasswords(user.HashedPassword, p)
}

func (user *User) SetPassword(p string) {
	user.HashedPassword = hashAndSalt(p)
}

func getUUID() string {
	return uuid.Must(uuid.NewV4()).String()
}

func (user *User) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("PaymentID", getUUID())
}

func FindUser(db *gorm.DB, user string, pass string) (*User, bool, error) {
	u := &User{}
	query := db.Where("email = ?", user).Take(u)
	if query.RecordNotFound() {
		return nil, false, query.Error
	}

	if u.MatchPassword(pass) {
		return u, true, nil
	}
	return nil, true, errors.New("wrong password")

}
