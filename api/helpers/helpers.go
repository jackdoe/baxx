package helpers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/notification"
	"github.com/jackdoe/baxx/user"
	"github.com/jinzhu/gorm"
)

func GetUserStatus(db *gorm.DB, u *user.User) (*common.UserStatusOutput, error) {
	tokens, err := ListTokens(db, u)
	if err != nil {
		return nil, err
	}
	tokensTransformed := []*common.TokenOutput{}
	for _, t := range tokens {
		usedSize, usedInodes, err := file.GetQuotaUsed(db, t)
		if err != nil {
			return nil, err
		}
		rules, err := ListNotifications(db, t)
		if err != nil {
			return nil, err
		}

		tokensTransformed = append(tokensTransformed, TransformTokenForSending(t, usedSize, usedInodes, rules))
	}

	used := uint64(0)
	for _, t := range tokens {
		used += t.SizeUsed
	}

	vl := &user.VerificationLink{}
	db.Where("email = ?", u.Email).Last(vl)

	return &common.UserStatusOutput{
		UserID:                u.ID,
		UsedSize:              used,
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
	}, nil
}

func TransformTokenForSending(t *file.Token, usedSize, usedInodes int64, rules []*notification.NotificationRule) *common.TokenOutput {
	ru := []common.NotificationRuleOutput{}
	for _, r := range rules {
		ru = append(ru, notification.TransformRuleToOutput(r))
	}
	return &common.TokenOutput{
		ID:                t.ID,
		UUID:              t.UUID,
		Name:              t.Name,
		WriteOnly:         t.WriteOnly,
		NumberOfArchives:  t.NumberOfArchives,
		CreatedAt:         t.CreatedAt,
		QuotaUsed:         uint64(usedSize),
		QuotaInodeUsed:    uint64(usedInodes),
		Quota:             t.Quota,
		QuotaInode:        t.QuotaInode,
		NotificationRules: ru,
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

func CreateToken(db *gorm.DB, writeOnly bool, u *user.User, bucket string, numOfArchives uint64, name string, quota uint64, quotaInode uint64, maxTokens uint64) (*file.Token, error) {
	t := &file.Token{
		UUID:             common.GetUUID(),
		Salt:             strings.Replace(common.GetUUID(), "-", "", -1),
		Bucket:           bucket,
		UserID:           u.ID,
		WriteOnly:        writeOnly,
		NumberOfArchives: numOfArchives,
		Name:             name,
		Quota:            quota,
		QuotaInode:       quotaInode,
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

func CreateTokenAndNotification(s *file.Store, db *gorm.DB, u *user.User, bucket string, writeOnly bool, numOfArchives uint64, name string, quota uint64, quotaInode uint64, maxTokens uint64, n common.CreateNotificationInput) (*file.Token, error) {
	t, err := CreateToken(db, writeOnly, u, bucket, numOfArchives, name, quota, quotaInode, maxTokens)
	if err != nil {
		return nil, err
	}
	n.TokenUUID = t.UUID
	_, err = CreateNotificationRule(db, u, &n)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func CreateNotificationRule(db *gorm.DB, u *user.User, json *common.CreateNotificationInput) (*notification.NotificationRule, error) {
	token, err := FindTokenForUser(db, json.TokenUUID, u)
	if err != nil {
		return nil, err
	}

	_, err = regexp.Compile(json.Regexp)
	if err != nil {
		return nil, err
	}

	if json.AcceptableAgeDays < 0 {
		return nil, fmt.Errorf("age_days must be >= 0 (0 is disabled), got: %d", json.AcceptableAgeDays)
	}

	if json.AcceptableSizeDeltaPercentBetweenVersions < 0 {
		return nil, fmt.Errorf("delta_percent must be >= 0 (0 is disabled), got: %d", json.AcceptableSizeDeltaPercentBetweenVersions)
	}

	if json.AcceptableSizeDeltaPercentBetweenVersions == 0 && json.AcceptableAgeDays == 0 {
		return nil, fmt.Errorf("both ")
	}
	n := &notification.NotificationRule{
		TokenID:                                   token.ID,
		UserID:                                    u.ID,
		Regexp:                                    json.Regexp,
		Name:                                      json.Name,
		UUID:                                      common.GetUUID(),
		AcceptableAgeSeconds:                      uint64(json.AcceptableAgeDays) * 86400,
		AcceptableSizeDeltaPercentBetweenVersions: uint64(json.AcceptableSizeDeltaPercentBetweenVersions),
	}

	if err := db.Create(n).Error; err != nil {
		return nil, err
	}
	return n, nil
}

func ChangeNotificationRule(db *gorm.DB, u *user.User, json *common.ModifyNotificationInput) (*notification.NotificationRule, error) {
	n := &notification.NotificationRule{}
	if err := db.Where("uuid = ? AND user_id = ?", json.UUID, u.ID).First(&n).Error; err != nil {
		return nil, err
	}

	if json.Name != nil {
		n.Name = *json.Name
	}
	if json.AcceptableSizeDeltaPercentBetweenVersions != nil {
		if *json.AcceptableSizeDeltaPercentBetweenVersions < 0 {
			return nil, fmt.Errorf("delta_percent must be >= 0 (0 is disabled), got: %d", *json.AcceptableSizeDeltaPercentBetweenVersions)
		}

		n.AcceptableSizeDeltaPercentBetweenVersions = uint64(*json.AcceptableSizeDeltaPercentBetweenVersions)
	}
	if json.AcceptableAgeDays != nil {
		if *json.AcceptableAgeDays < 0 {
			return nil, fmt.Errorf("age_days must be >= 0 (0 is disabled), got: %d", *json.AcceptableAgeDays)
		}

		n.AcceptableAgeSeconds = uint64(*json.AcceptableAgeDays) * 86400
	}
	if json.Regexp != nil {
		n.Regexp = *json.Regexp
	}

	_, err := regexp.Compile(n.Regexp)
	if err != nil {
		return nil, err
	}

	if err := db.Save(n).Error; err != nil {
		return nil, err
	}
	return n, nil
}

func ListNotifications(db *gorm.DB, t *file.Token) ([]*notification.NotificationRule, error) {
	rules := []*notification.NotificationRule{}
	if err := db.Where("token_id = ?", t.ID).Find(&rules).Error; err != nil {
		return nil, err
	}

	return rules, nil
}
