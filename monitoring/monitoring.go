package monitoring

import (
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type MonitoringPerNode struct {
	ID             uint64 `gorm:"primary_key"`
	Kind           string `gorm:"type:varchar(255) not null"`
	NodeID         string `gorm:"type:varchar(255) not null"`
	AlertText      string `gorm:"type:text not null"`
	Schedule       float64
	Tick           time.Time
	NotifiedAt     *time.Time
	AcknowledgedAt *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (m *MonitoringPerNode) String() string {
	s := fmt.Sprintf("%s %s %s - last tick %s (%fs ago)", m.Kind, m.NodeID, m.AlertText, m.Tick.Format(time.ANSIC), time.Since(m.Tick).Seconds())
	if m.NotifiedAt != nil {
		s = fmt.Sprintf("%s, notified %fs ago", s, int(time.Since(*m.NotifiedAt).Seconds()))
	}

	if m.AcknowledgedAt != nil {
		s = fmt.Sprintf("%s, ack %fs ago", s, int(time.Since(*m.AcknowledgedAt).Seconds()))
	}

	return s
}

func Hostname() string {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return name
}

func MustInitNode(db *gorm.DB, kind string, text string, schedule float64) {
	err := InitNode(db, kind, text, schedule)
	if err != nil {
		log.Panic(err)
	}
}

func InitNode(db *gorm.DB, kind string, text string, schedule float64) error {
	n := &MonitoringPerNode{}
	if err := db.FirstOrCreate(&n, MonitoringPerNode{Kind: kind, NodeID: Hostname()}).Error; err != nil {
		return err
	}

	n.Schedule = schedule
	n.AlertText = text
	n.Tick = time.Now()
	return db.Save(n).Error
}
func MustTick(db *gorm.DB, kind string) {
	err := Tick(db, kind)
	if err != nil {
		log.Panic(err)
	}

}
func Tick(db *gorm.DB, kind string) error {
	n := &MonitoringPerNode{}
	if err := db.First(&n, MonitoringPerNode{Kind: kind, NodeID: Hostname()}).Error; err != nil {
		return err
	}

	n.Tick = time.Now()
	return db.Save(n).Error
}

const MIN_NOTIFICATION_TIME = float64(3600)

func Watch(db *gorm.DB) ([]*MonitoringPerNode, error) {
	items := []*MonitoringPerNode{}
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}

	out := []*MonitoringPerNode{}
	for _, item := range items {
		if time.Since(item.Tick).Seconds() > item.Schedule {
			// either there is no old notification
			// there is old notification but it is older than MIN_NOTIFICATION_TIME
			// there is acknowledgement that happened after the last notification,
			// but we still have a problem after that
			if item.NotifiedAt == nil ||
				time.Since(*item.NotifiedAt).Seconds() > MIN_NOTIFICATION_TIME ||
				(item.AcknowledgedAt != nil &&
					item.NotifiedAt != nil &&
					item.AcknowledgedAt.Sub(*item.NotifiedAt).Seconds() > 1) {
				out = append(out, item)
			}
		}
	}

	return out, nil
}
