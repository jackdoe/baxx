package user

import (
	"errors"
	"fmt"
	. "github.com/jackdoe/baxx/config"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
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

func (user *User) ListTokens(db *gorm.DB) ([]*Token, error) {
	tokens := []*Token{}
	if err := db.Where("user_id = ?", user.ID).Find(&tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}

func getUUID() string {
	return uuid.Must(uuid.NewV4()).String()
}

func (user *User) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("PaymentID", getUUID())
}

type Token struct {
	ID     uint64 `gorm:"primary_key"`
	UUID   string `gorm:"not null"`
	Salt   string `gorm:"not null;type:varchar(32)"`
	UserID uint64 `gorm:"not null"`

	WriteOnly        bool   `gorm:"not null"`
	NumberOfArchives uint64 `gorm:"not null"`
	SizeUsed         uint64 `gorm:"not null;default:0"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (user *User) GetQuotaLeft(db *gorm.DB) (int64, error) {
	used, err := user.GetQuotaUsed(db)
	if err != nil {
		return 0, err
	}
	return int64(user.Quota) - int64(used), nil
}

func (user *User) GetQuotaUsed(db *gorm.DB) (int64, error) {
	tokens, err := user.ListTokens(db)
	if err != nil {
		return 0, err
	}

	used := uint64(0)
	for _, t := range tokens {
		used += t.SizeUsed
	}
	return int64(used), nil
}

func (user *User) CreateToken(db *gorm.DB, writeOnly bool, numOfArchives uint64) (*Token, error) {
	t := &Token{
		UUID:             getUUID(),
		Salt:             strings.Replace(getUUID(), "-", "", -1),
		UserID:           user.ID,
		WriteOnly:        writeOnly,
		NumberOfArchives: numOfArchives,
	}

	tokens, err := user.ListTokens(db)
	if err != nil {
		return nil, err
	}

	if len(tokens) >= int(CONFIG.MaxTokens) {
		return nil, fmt.Errorf("max tokens created (max=%d)", CONFIG.MaxTokens)
	}

	if err := db.Create(t).Error; err != nil {
		return nil, err
	}

	return t, nil
}

func FindToken(db *gorm.DB, token string) (*Token, *User, error) {
	t := &Token{}

	query := db.Where("uuid = ?", token).Take(t)
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
