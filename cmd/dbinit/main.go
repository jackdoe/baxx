package main

import (
	"log"
	"os"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	al "github.com/jackdoe/baxx/api/action_log"
	"github.com/jackdoe/baxx/api/file"
	notification "github.com/jackdoe/baxx/api/notification_rules"
	"github.com/jackdoe/baxx/api/user"
	"github.com/jackdoe/baxx/message"
	"github.com/jackdoe/baxx/monitoring"
	"github.com/jinzhu/gorm"
)

func initDatabase(db *gorm.DB) {
	if err := db.AutoMigrate(
		&user.User{},
		&user.VerificationLink{},
		&file.Token{},
		&file.FileMetadata{},
		&file.FileVersion{},
		&al.ActionLog{},
		&user.PaymentHistory{},
		&notification.NotificationRule{},
		&notification.NotificationForFileVersion{},
		&notification.NotificationForQuota{},
		&monitoring.MonitoringPerNode{},
		&monitoring.DiskUsagePerNode{},
		&monitoring.DiskIOPerNode{},
		&monitoring.DiskMDPerNode{},
		&monitoring.MemStatsPerNode{},
		&message.EmailQueueItem{},
	).Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&message.EmailQueueItem{}).AddIndex("idx_email_sent", "sent").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&user.VerificationLink{}).AddUniqueIndex("idx_user_sent_at", "user_id", "sent_at").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&monitoring.DiskUsagePerNode{}).AddIndex("idx_monitoring_du_node_kind_time", "node_id", "kind", "created_at").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&monitoring.DiskIOPerNode{}).AddIndex("idx_monitoring_io_node_kind_time", "node_id", "kind", "created_at").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&monitoring.DiskMDPerNode{}).AddIndex("idx_monitoring_md_node_kind_time", "node_id", "kind", "created_at").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&monitoring.MemStatsPerNode{}).AddIndex("idx_monitoring_mem_node_time", "node_id", "created_at").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&monitoring.MonitoringPerNode{}).AddUniqueIndex("idx_monitoring_kind_node_id", "kind", "node_id").Error; err != nil {
		log.Panic(err)
	}

	// not unique index, we can have many links for same email, they could expire
	if err := db.Model(&user.VerificationLink{}).AddIndex("idx_vl_email", "email").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&user.User{}).AddUniqueIndex("idx_payment_id", "payment_id").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&file.FileVersion{}).AddIndex("idx_token_sha", "token_id", "sha256").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&file.FileVersion{}).AddIndex("idx_fv_metadata", "file_metadata_id").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&file.FileMetadata{}).AddUniqueIndex("idx_fm_token_id_path_2", "token_id", "path", "filename").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&notification.NotificationRule{}).AddIndex("idx_n_user_id_token_id", "user_id", "token_id").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&notification.NotificationForFileVersion{}).AddUniqueIndex("idx_nfv_rule_fv", "file_version_id", "notification_rule_id").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&notification.NotificationForQuota{}).AddUniqueIndex("idx_nfq_token", "token_id").Error; err != nil {
		log.Panic(err)
	}
}

func main() {
	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Panic(err)
	}
	db.LogMode(true)
	defer db.Close()

	initDatabase(db)
}
