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
	fmt.Println(Render(HelpObject{Template: EmailPaymentPlease, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: EmailValidation, Email: status.Email, Status: status}))
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
	fmt.Println(Render(HelpObject{Template: TokenMeta, Email: status.Email, Status: status}))
	fmt.Println(Render(HelpObject{Template: SyncMeta, Email: status.Email, Status: status}))

	fmt.Println(Render(HelpObject{Template: EmailNotification, Email: status.Email, Status: status, Notifications: []common.FileNotification{
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
	}}))

}
