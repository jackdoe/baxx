package main

import (
	"fmt"
	baxx "github.com/jackdoe/baxx/client"
	bcommon "github.com/jackdoe/baxx/common"
	. "github.com/jackdoe/baxx/help"
	"github.com/marcusolsson/tui-go"
	"log"
	"time"
)

var logo = `██████╗  █████╗ ██╗  ██╗██╗  ██╗
██╔══██╗██╔══██╗╚██╗██╔╝╚██╗██╔╝
██████╔╝███████║ ╚███╔╝  ╚███╔╝
██╔══██╗██╔══██║ ██╔██╗  ██╔██╗
██████╔╝██║  ██║██╔╝ ██╗██╔╝ ██╗
╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝`

func popup(ui tui.UI, root *tui.Box, closeIsExit bool, onClose func(), title string, msg ...string) {
	text := tui.NewVBox()

	scroll := tui.NewScrollArea(text)
	closeLabel := "[ Back ]"
	if closeIsExit {
		closeLabel = "[ Exit ]"
	}
	close := tui.NewButton(closeLabel)

	close.SetFocused(true)
	close.SetSizePolicy(tui.Preferred, tui.Maximum)
	padder := tui.NewPadder(1, 1, scroll)
	p := tui.NewVBox(
		padder,
		close,
	)
	p.SetBorder(true)
	p.SetTitle(fmt.Sprintf("baxx.dev - %s", title))
	p.SetSizePolicy(tui.Maximum, tui.Maximum)
	bye := func() {
		ui.ClearKeybindings()
		ui.SetKeybinding("Esc", func() { ui.Quit() })
		ui.SetWidget(root)
		if onClose != nil {
			onClose()
		}
		if closeIsExit {
			ui.Quit()
		}
	}

	close.OnActivated(func(b *tui.Button) {
		bye()
	})

	ui.ClearKeybindings()
	ui.SetKeybinding("Up", func() { scroll.Scroll(0, -1) })
	ui.SetKeybinding("Down", func() { scroll.Scroll(0, 1) })
	ui.SetKeybinding("k", func() { scroll.Scroll(0, -1) })
	ui.SetKeybinding("j", func() { scroll.Scroll(0, 1) })
	ui.SetKeybinding("Esc", func() { bye() })
	ui.SetWidget(p)

	for _, m := range msg {
		l := tui.NewLabel(m)
		l.SetSizePolicy(tui.Maximum, tui.Minimum)
		l.SetWordWrap(false)
		text.Append(l)
	}
}
func apiError(err error) string {
	return fmt.Sprintf(`
API Error:

  %s

please contact help@baxx.dev if it persists`, err.Error())
}

func registrationForm(ui tui.UI, bc *baxx.Client, onRegister func(string, string)) *tui.Box {
	user := tui.NewEntry()
	user.SetFocused(true)

	password := tui.NewEntry()
	password.SetEchoMode(tui.EchoModePassword)

	confirmPassword := tui.NewEntry()
	confirmPassword.SetEchoMode(tui.EchoModePassword)

	form := tui.NewGrid(0, 0)
	form.AppendRow(tui.NewLabel("Email"))
	form.AppendRow(user)
	form.AppendRow(tui.NewSpacer())
	form.AppendRow(tui.NewLabel("Password"))
	form.AppendRow(password)
	form.AppendRow(tui.NewSpacer())
	form.AppendRow(tui.NewLabel("Confirm Password"))
	form.AppendRow(confirmPassword)

	register := tui.NewButton("[Register]")

	quit := tui.NewButton("[Quit]")
	help := tui.NewButton("[Help]")
	pitch := tui.NewButton("[What/Why/How]")
	tos := tui.NewButton("[Terms Of Service]")
	buttonsRegister := tui.NewHBox(
		tui.NewSpacer(),
		tui.NewPadder(1, 0, register),
		tui.NewSpacer(),
	)

	buttonsTop := tui.NewHBox(
		tui.NewSpacer(),
		tui.NewPadder(1, 0, help),
		tui.NewPadder(1, 0, pitch),
		tui.NewPadder(1, 0, tos),
		tui.NewSpacer(),
	)

	buttonsBottom := tui.NewHBox(
		tui.NewSpacer(),
		tui.NewPadder(1, 0, quit),
		tui.NewSpacer(),
	)

	window := tui.NewVBox(
		tui.NewPadder(1, 1, tui.NewLabel(logo)),
		tui.NewPadder(1, 0, tui.NewLabel(Render(GUI_INTRO, nil))),
		tui.NewPadder(1, 1, form),
		tui.NewPadder(1, 0, tui.NewLabel("Registering means you agree with\nthe terms of service!")),
		tui.NewPadder(1, 0, tui.NewLabel("")),
		buttonsRegister,
		tui.NewPadder(1, 0, tui.NewLabel("")),

		buttonsTop,
		tui.NewPadder(1, 0, tui.NewLabel("")),
		buttonsBottom,
	)
	window.SetBorder(true)

	wrapper := tui.NewVBox(
		tui.NewSpacer(),
		window,
		tui.NewSpacer(),
	)

	content := tui.NewHBox(tui.NewSpacer(), wrapper, tui.NewSpacer())

	root := content
	chain := &tui.SimpleFocusChain{}
	chain.Set(user, password, confirmPassword, register, help, pitch, tos, quit)
	ui.SetFocusChain(chain)

	quit.OnActivated(func(b *tui.Button) {
		ui.Quit()
	})

	help.OnActivated(func(b *tui.Button) {
		popup(ui, root, false, nil, "HELP", Render(EMAIL_AFTER_REGISTRATION, bcommon.EMPTY_STATUS))
	})

	pitch.OnActivated(func(b *tui.Button) {
		popup(ui, root, false, nil, "HELP", Render(PITCH, nil))
	})

	tos.OnActivated(func(b *tui.Button) {
		popup(ui, root, false, nil, "Terms Of Service", Render(TERMS_AND_CONDITIONS, nil))
	})

	register.OnActivated(func(b *tui.Button) {
		p1 := password.Text()
		p2 := confirmPassword.Text()
		email := user.Text()

		if p1 != p2 {
			popup(ui, root, false, nil, "ERROR", "passwords must match")
			return
		}

		if p1 == "" {
			popup(ui, root, false, nil, "ERROR", Render(GUI_PASS_REQUIRED, nil))
			return
		}
		if email == "" {
			popup(ui, root, false, nil, "ERROR", Render(GUI_EMAIL_REQUIRED, nil))
			return
		}
		_, err := bc.Status(&bcommon.CreateUserInput{Email: email, Password: p1})
		if err == nil {
			onRegister(email, p1)
		} else {
			_, err := bc.Register(&bcommon.CreateUserInput{Email: email, Password: p1})
			if err != nil {
				popup(ui, root, false, nil, "ERROR", apiError(err))
			} else {
				onRegister(email, p1)
			}
		}
	})

	return root
}

func postRegistration(ui tui.UI, bc *baxx.Client, email, pass string) *tui.Box {
	quit := tui.NewButton("[Quit]")
	resend := tui.NewButton("[Resend Verification Email]")
	help := tui.NewButton("[Help]")
	timer := tui.NewLabel("")
	buttonsBottom := tui.NewHBox(
		tui.NewSpacer(),
		tui.NewPadder(1, 0, help),
		tui.NewPadder(1, 0, resend),
		tui.NewPadder(1, 0, quit),
		tui.NewSpacer(),
	)

	emailVerified := tui.NewLabel("[ ... ]")
	subscribed := tui.NewLabel("[ ... ]")

	window := tui.NewVBox(
		tui.NewPadder(1, 1, tui.NewLabel(logo)),
		tui.NewPadder(1, 0, tui.NewLabel("")),
		tui.NewPadder(1, 0, tui.NewLabel("Email: "+email)),
		tui.NewPadder(1, 0, emailVerified),
		tui.NewPadder(1, 0, tui.NewLabel("")),
		tui.NewPadder(1, 0, tui.NewLabel("Subscription:")),
		tui.NewPadder(1, 0, subscribed),
		tui.NewPadder(1, 0, tui.NewLabel("                                                                        ")),
		tui.NewPadder(1, 0, timer),
		tui.NewPadder(1, 0, tui.NewLabel("")),
		buttonsBottom,
	)
	window.SetBorder(true)

	wrapper := tui.NewVBox(
		tui.NewSpacer(),
		window,
		tui.NewSpacer(),
	)

	content := tui.NewHBox(tui.NewSpacer(), wrapper, tui.NewSpacer())
	chain := &tui.SimpleFocusChain{}
	chain.Set(help, resend, quit)
	ui.SetFocusChain(chain)
	stop := make(chan bool, 1)
	quit.OnActivated(func(b *tui.Button) {
		ui.Quit()
	})

	help.OnActivated(func(b *tui.Button) {
		popup(ui, content, false, nil, "HELP", Render(PITCH, nil), "", Render(EMAIL_AFTER_REGISTRATION, bcommon.EMPTY_STATUS))
	})

	refreshStatus := func() error {
		status, err := bc.Status(&bcommon.CreateUserInput{Email: email, Password: pass})
		if err != nil {
			return err
		}

		if status.EmailVerified != nil {
			emailVerified.SetText("Verified at " + status.EmailVerified.Format(time.ANSIC))
		} else {
			emailVerified.SetText("Verification pending.\nPlease check your spam folder.")
		}
		if status.Paid {
			subscribed.SetText("Active")
		} else {
			subscribed.SetText("Activate at https://baxx.dev/sub/" + status.PaymentID)
		}
		if status.Paid && status.EmailVerified != nil {
			stop <- true
			popup(ui, content, true, nil, "SUCCESS", "Your account is now ready to be used", "", "", Render(EMAIL_AFTER_REGISTRATION, status))
		}
		return nil
	}

	resendEmail := func() error {
		_, err := bc.ReplaceEmail(&bcommon.CreateUserInput{Email: email, Password: pass}, email)
		return err
	}

	work := make(chan func() error, 1)

	resend.OnActivated(func(b *tui.Button) {
		work <- resendEmail
	})

	go func() {
		spinner := []string{"/", "-", "\\", "|"}
		i := 0
		for {
			select {
			case cb := <-work:
				err := cb()
				if err != nil {
					wait := make(chan bool)
					popup(ui, content, false, func() { wait <- true }, "ERROR", apiError(err))
					ui.Update(func() {})
					<-wait
				}
				timer.SetText("Refreshing.. " + spinner[i%len(spinner)])
				ui.Update(func() {})
				time.Sleep(1 * time.Second)
				i++
			case <-stop:
				return
			}
		}
	}()

	go func() {
		for {
			work <- refreshStatus
			time.Sleep(1 * time.Second)
		}
	}()

	return content
}

func main() {
	statusUpdate := make(chan string)

	bc := baxx.NewClient(nil, "https://baxx.dev", statusUpdate)
	status := tui.NewStatusBar("")
	go func() {
		for {
			s := <-statusUpdate
			status.SetText(s)
		}
	}()

	content := tui.NewHBox()
	root := tui.NewVBox(
		content,
		status,
	)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

	//ui.SetWidget(tui.NewVBox(
	//	postRegistration(ui, bc, "jack@sofialondonmoskva.com", "asdasdasdasd"),
	//	status,
	//))

	register := registrationForm(ui, bc, func(u, p string) {
		ui.SetWidget(tui.NewVBox(
			postRegistration(ui, bc, u, p),
			status,
		))
	})
	content.Prepend(register)

	ui.SetKeybinding("Esc", func() { ui.Quit() })

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}
