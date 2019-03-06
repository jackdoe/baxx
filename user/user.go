package user

import (
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
	ID        uint64    `gorm:"primary_key" json:"-"`
	UserID    uint64    `gorm:"not null" json:"-"`
	IPN       string    `gorm:"not null;type:text" json:"-"`
	IPNRAW    string    `gorm:"not null;type:text" json:"-"`
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
	ID                    uint64     `gorm:"primary_key" json:"-"`
	PaymentID             string     `gorm:"not null" json:"-"`
	Email                 string     `gorm:"not null" json:"-"`
	Quota                 uint64     `gorm:"not null;default:10737418240" json:"quota"`
	QuotaUsed             uint64     `gorm:"not null;default:0" json:"quota_used"`
	EmailVerified         *time.Time `json:"-"`
	StartedSubscription   *time.Time `json:"-"`
	CancelledSubscription *time.Time `json:"-"`
	HashedPassword        string     `gorm:"not null" json:"-"`
	CreatedAt             time.Time  `json:"-"`
	UpdatedAt             time.Time  `json:"-"`
}

type NotificationRule struct {
	ID     uint64 `gorm:"primary_key" json:"-"`
	UserID uint64 `gorm:"not null" json:"-"`
	Name   string `gorm:"not null" json:"-"`
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

func (user *User) ListTokens(db *gorm.DB) ([]*Token, error) {
	tokens := []*Token{}
	if err := db.Where("user_id = ?", user.ID).Find(&tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}

func getUUID() string {
	return fmt.Sprintf("%s", uuid.Must(uuid.NewV4()))
}

func (user *User) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("PaymentID", getUUID())
	return nil
}

type Token struct {
	ID     string `gorm:"primary_key"  json:"token"`
	Salt   string `gorm:"not null";type:"varchar(32)" json:"-"`
	UserID uint64 `gorm:"not null" json:"-"`

	WriteOnly        bool   `gorm:"not null" json:"write_only"`
	NumberOfArchives uint64 `gorm:"not null" json:"keep_n_versions"`
	SizeUsed         uint64 `gorm:"not null;default:0" json:"size_used"`
	CreatedAt        time.Time
	UpdatedAt        time.Time `json:"-"`
}

func (token *Token) BeforeCreate(scope *gorm.Scope) error {
	id := uuid.Must(uuid.NewV4())

	scope.SetColumn("ID", fmt.Sprintf("%s", id))
	return nil
}

func FindToken(db *gorm.DB, token string) (*Token, *User, error) {
	t := &Token{}

	query := db.Where("id = ?", token).Take(t)
	if query.RecordNotFound() {
		return nil, nil, query.Error
	}

	u := &User{}
	query = db.Where("id = ?", t.UserID).Take(u)
	if query.RecordNotFound() {
		return nil, nil, query.Error
	}

	return t, u, nil
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
