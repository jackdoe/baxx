package helpers

import (
	"fmt"
	"strings"

	"github.com/jackdoe/baxx/api/file"
	"github.com/jackdoe/baxx/api/user"
	"github.com/jackdoe/baxx/common"

	"github.com/jinzhu/gorm"
)

func GetUserStatus(db *gorm.DB, u *user.User) (*common.UserStatusOutput, error) {
	tokens, err := ListTokens(db, u)
	if err != nil {
		return nil, err
	}
	tokensTransformed := []*common.TokenOutput{}
	for _, t := range tokens {
		tokensTransformed = append(tokensTransformed, TransformTokenForSending(t))
	}

	usedSize := uint64(0)
	usedInodes := uint64(0)
	for _, t := range tokens {
		usedSize += t.SizeUsed
		usedInodes += t.CountFiles
	}

	vl := &user.VerificationLink{}
	db.Where("email = ?", u.Email).Last(vl)

	return &common.UserStatusOutput{
		UserID:                u.ID,
		EmailVerified:         u.EmailVerified,
		StartedSubscription:   u.StartedSubscription,
		CancelledSubscription: u.CancelledSubscription,
		Tokens:                tokensTransformed,
		SubscribeURL:          fmt.Sprintf("https://baxx.dev/sub/%s", u.PaymentID),
		CancelSubscriptionURL: fmt.Sprintf("https://baxx.dev/unsub/%s", u.PaymentID),
		LastVerificationID:    vl.ID,
		Paid:                  u.Paid(),
		PaymentID:             u.PaymentID,
		Email:                 u.Email,
		QuotaUsed:             uint64(usedSize),
		QuotaInodeUsed:        uint64(usedInodes),
		Quota:                 u.Quota,
		QuotaInode:            u.QuotaInode,
	}, nil
}

func TransformTokenForSending(t *file.Token) *common.TokenOutput {
	return &common.TokenOutput{
		ID:         t.ID,
		UUID:       t.UUID,
		Name:       t.Name,
		WriteOnly:  t.WriteOnly,
		CreatedAt:  t.CreatedAt,
		SizeUsed:   t.SizeUsed,
		InodesUsed: t.CountFiles,
	}
}

func ListTokens(db *gorm.DB, u *user.User) ([]*file.Token, error) {
	tokens := []*file.Token{}
	if err := db.Where("user_id = ?", u.ID).Find(&tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}

func FindTokenForUser(db *gorm.DB, token string, u *user.User) (*file.Token, error) {
	t := &file.Token{}

	query := db.Where("uuid = ? AND user_id = ?", token, u.ID).Take(t)
	if query.Error != nil {
		return nil, query.Error
	}

	return t, nil
}

func StopNotifications(db *gorm.DB, u *user.User, fmid uint64) (*file.FileMetadata, error) {
	var fm file.FileMetadata
	if err := db.Where("id = ?", fmid).Take(&fm).Error; err != nil {
		return nil, err
	}
	var t file.Token
	if err := db.Where("id = ? AND user_id = ?", fm.TokenID, u.ID).Take(&t).Error; err != nil {
		return nil, err
	}

	fm.AcceptableAge = 0
	fm.AcceptableDelta = 0
	if err := db.Save(fm).Error; err != nil {
		return nil, err
	}
	return &fm, nil
}

func CreateToken(db *gorm.DB, u *user.User, writeOnly bool, name string, maxTokens uint64) (*file.Token, error) {
	t := &file.Token{
		UUID:      common.GetUUID(),
		Salt:      strings.Replace(common.GetUUID(), "-", "", -1),
		UserID:    u.ID,
		WriteOnly: writeOnly,
		Name:      name,
	}

	tokens, err := ListTokens(db, u)
	if err != nil {
		return nil, err
	}

	if len(tokens) >= int(maxTokens) {
		return nil, fmt.Errorf("max tokens created (max=%d)", maxTokens)
	}

	if err := db.Create(t).Error; err != nil {
		return nil, err
	}

	return t, nil
}

func FindTokenAndUser(db *gorm.DB, token string) (*file.Token, *user.User, error) {
	t := &file.Token{}

	query := db.Where("uuid = ?", token).Take(t)
	if query.RecordNotFound() {
		return nil, nil, query.Error
	}

	u := &user.User{}
	query = db.Where("id = ?", t.UserID).Take(u)
	if query.RecordNotFound() {
		return nil, nil, query.Error
	}

	return t, u, nil
}
