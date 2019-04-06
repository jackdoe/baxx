package user

import (
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"

	"github.com/jackdoe/baxx/common"
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
	Email      string     `gorm:"type:varchar(255) not null" json:"-"`
	VerifiedAt *time.Time `gorm:"null" json:"-"`
	SentAt     uint64     `gorm:"not null" json:"-"`
	UpdatedAt  time.Time  `json:"-"`
	CreatedAt  time.Time  `json:"-"`
}

type User struct {
	ID                    uint64     `gorm:"primary_key" json:"-"`
	PaymentID             string     `gorm:"not null" json:"-"`
	Email                 string     `gorm:"not null;unique" json:"-"`
	EmailVerified         *time.Time `json:"-"`
	StartedSubscription   *time.Time `json:"-"`
	CancelledSubscription *time.Time `json:"-"`
	Quota                 uint64     `gorm:"not null;default:10737418240" json:"quota"`
	QuotaInode            uint64     `gorm:"not null;default:1000" json:"quota_inode"`
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
		ID:     common.GetUUID(),
		UserID: user.ID,
		Email:  user.Email,
		SentAt: uint64(time.Now().UnixNano() / 1e6),
	}
}

func (user *User) MatchPassword(p string) bool {
	return comparePasswords(user.HashedPassword, p)
}

func (user *User) SetPassword(p string) {
	user.HashedPassword = hashAndSalt(p)
}

func (user *User) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("PaymentID", common.GetUUID())
}

func FindUser(db *gorm.DB, user string, pass string) (*User, error) {
	u := &User{}
	if err := db.Where("email = ?", user).Take(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func Exists(db *gorm.DB, user string) bool {
	u := &User{}
	q := db.Where("email = ?", user).Take(u)
	return !q.RecordNotFound()
}
