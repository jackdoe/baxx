package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func hashAndSalt(pwd string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	return string(hash)
}

func comparePasswords(hashedPwd string, plainPwd string) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(plainPwd))
	if err != nil {
		return false
	}

	return true
}

type PaymentHistory struct {
	ID        string    `gorm:"primary_key" json:"-"`
	UserID    uint64    `gorm:"not null" json:"-"`
	IPN       string    `gorm:"not null";type:"text" json:"-"`
	IPNRAW    string    `gorm:"not null";type:"text" json:"-"`
	UpdatedAt time.Time `json:"-"`
	CreatedAt time.Time `json:"-"`
}

type VerificationLink struct {
	ID         string     `gorm:"primary_key" json:"-"`
	UserID     uint64     `gorm:"not null" json:"-"`
	Email      string     `gorm:"not null" json:"-"`
	VerifiedAt *time.Time `gorm:"null" json:"-"`
	SentAt     uint64     `gorm:"not null" json:"-"`
	UpdatedAt  time.Time  `json:"-"`
	CreatedAt  time.Time  `json:"-"`
}

type User struct {
	ID               uint64     `gorm:"primary_key" json:"-"`
	SemiSecretID     string     `gorm:"not null" json:"secret"`
	Seed             string     `gorm:"not null" json:"-"`
	PaymentID        string     `gorm:"not null" json:"-"`
	Email            string     `gorm:"not null" json:"-"`
	EmailVerified    *time.Time `json:"-"`
	PaidSubscription *time.Time `json:"-"`
	HashedPassword   string     `gorm:"not null" json:"-"`
	CreatedAt        time.Time  `json:"-"`
	UpdatedAt        time.Time  `json:"-"`
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
	return fmt.Sprintf("%s", uuid.Must(uuid.NewV4()))
}

func (user *User) SetSemiSecretID() {
	user.SemiSecretID = getUUID()
}

func (user *User) BeforeSave(scope *gorm.Scope) error {
	scope.SetColumn("Seed", getUUID())
	return nil
}

func (user *User) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("SemiSecretID", getUUID())
	scope.SetColumn("Seed", getUUID())
	scope.SetColumn("PaymentID", getUUID())
	return nil
}

/* do some validataion */

type NotificationDest struct {
	Type  string `binding:"required"`
	Value string `binding:"required"`
}

type NotificationRule struct {
	Match        string `binding:"required"`
	Type         string `binding:"required"`
	Value        int64
	Destinations []NotificationDest
}

type NotificationConfiguration struct {
	Rules []*NotificationRule `binding:"required"`
}

type Token struct {
	ID                 string    `gorm:"primary_key"  json:"token"`
	Salt               string    `gorm:"not null";type:"varchar(32)" json:"-"`
	UserID             uint64    `gorm:"not null" json:"-"`
	WriteOnly          bool      `gorm:"not null" json:"-"`
	NumberOfArchives   uint64    `gorm:"not null" json:"-"`
	NotificationConfig string    `gorm:"not null";type:"text" json:"-"`
	CreatedAt          time.Time `json:"-"`
	UpdatedAt          time.Time `json:"-"`
}

func (token *Token) GetNotificationConfig() (*NotificationConfiguration, error) {
	nc := &NotificationConfiguration{}
	err := json.Unmarshal([]byte(token.NotificationConfig), nc)
	if err != nil {
		return nil, err
	}
	return nc, nil
}

func (token *Token) SetNotificationConfig(nc *NotificationConfiguration) error {
	b, err := json.Marshal(nc)
	if err != nil {
		return err
	}
	token.NotificationConfig = string(b)
	return nil
}

func (token *Token) BeforeCreate(scope *gorm.Scope) error {
	id := uuid.Must(uuid.NewV4())

	scope.SetColumn("ID", fmt.Sprintf("%s", id))
	return nil
}

func FindToken(db *gorm.DB, userSemiSecretId string, token string) (*Token, error) {
	t := &Token{}
	u := &User{}
	query := db.Where("semi_secret_id = ?", userSemiSecretId).Take(u)
	if query.RecordNotFound() {
		return nil, query.Error
	}

	query = db.Where("user_id = ? AND id = ?", u.ID, token).Take(t)
	if query.RecordNotFound() {
		return nil, query.Error
	}
	return t, nil
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
