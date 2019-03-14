package help

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackdoe/baxx/common"
)

func TestAll(t *testing.T) {
	status := common.EMPTY_STATUS

	fmt.Println(Render(HelpObject{Template: EmailAfterRegistration, Email: status.Email, Status: status}))

	fmt.Println(Render(HelpObject{Template: EmailPaymentCancel, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: EmailPaymentThanks, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: EmailValidation, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: FileDelete, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: FileDownload, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: FileList, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: FileUpload, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: FileMeta, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: GuiEmailRequired, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: GuiInfo, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: GuiPassRequired, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: GuiPitch, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: GuiTos, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: HtmlLinkError, Email: status.Email, Status: status, Err: errors.New("example error")}))
	fmt.Println(Render(HelpObject{Template: HtmlLinkExpired, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: HtmlVerificationOk, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: HtmlWaitPaypal, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: TokenCreate, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: TokenDelete, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: TokenList, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: TokenModify, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: TokenMeta, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: SyncShaMany, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: SyncShaOne, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: SyncMeta, Email: status.Email, Status: status}))

	fmt.Println(Render(HelpObject{Template: EmailNotification, Email: status.Email, Status: status, Notifications: []common.PerRuleGroup{
		common.PerRuleGroup{
			PerFile: []common.FileNotification{
				common.FileNotification{
					Age:             &common.AgeNotification{Overdue: 1 * time.Second, ActualAge: 1 * time.Second},
					Size:            &common.SizeNotification{PreviousSize: 5555, Delta: 0.123123},
					CreatedAt:       time.Now(),
					FullPath:        "/tmp/example.txt",
					LastVersionSize: 1,
				},
				common.FileNotification{
					Age:             &common.AgeNotification{Overdue: 1 * time.Second, ActualAge: 1 * time.Second},
					Size:            &common.SizeNotification{PreviousSize: 5555, Delta: 0.123123},
					CreatedAt:       time.Now(),
					FullPath:        "/tmp/example.txt",
					LastVersionSize: 1,
				},
				common.FileNotification{
					Age:             &common.AgeNotification{Overdue: 1 * time.Second, ActualAge: 1 * time.Second},
					Size:            &common.SizeNotification{PreviousSize: 5555, Delta: 0.123123},
					CreatedAt:       time.Now(),
					FullPath:        "/tmp/example.txt",
					LastVersionSize: 1,
				},
			},
			Rule: common.NotificationRuleOutput{
				Name:              "more than 1 day old database backup",
				Regexp:            "\\.sql",
				AcceptableAgeDays: 1,
				UUID:              "NOTIFICATION-UUID",
			},
		},

		common.PerRuleGroup{
			PerFile: []common.FileNotification{
				common.FileNotification{
					Age:             &common.AgeNotification{Overdue: 1 * time.Second, ActualAge: 1 * time.Second},
					Size:            &common.SizeNotification{PreviousSize: 5555, Delta: 0.123123},
					CreatedAt:       time.Now(),
					FullPath:        "/tmp/example.txt",
					LastVersionSize: 1,
				},
				common.FileNotification{
					Age:             &common.AgeNotification{Overdue: 1 * time.Second, ActualAge: 1 * time.Second},
					Size:            &common.SizeNotification{PreviousSize: 5555, Delta: 0.123123},
					CreatedAt:       time.Now(),
					FullPath:        "/tmp/example.txt",
					LastVersionSize: 1,
				},
				common.FileNotification{
					Age:             &common.AgeNotification{Overdue: 1 * time.Second, ActualAge: 1 * time.Second},
					Size:            &common.SizeNotification{PreviousSize: 5555, Delta: 0.123123},
					CreatedAt:       time.Now(),
					FullPath:        "/tmp/example.txt",
					LastVersionSize: 1,
				},
				common.FileNotification{
					Age:             &common.AgeNotification{Overdue: 1 * time.Second, ActualAge: 1 * time.Second},
					Size:            &common.SizeNotification{PreviousSize: 5555, Delta: 0.123123},
					CreatedAt:       time.Now(),
					FullPath:        "/tmp/example.txt",
					LastVersionSize: 1,
				},
			},
			Rule: common.NotificationRuleOutput{
				Name:              "more than 1 day old database backup",
				Regexp:            "\\.sql",
				AcceptableAgeDays: 1,
				UUID:              "NOTIFICATION-UUID",
			},
		},
	}}))

}
