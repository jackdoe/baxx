package main

import (
	"fmt"
	"log"
	"os"
	"time"

	baxx "github.com/jackdoe/baxx/client"
	bcommon "github.com/jackdoe/baxx/common"
	. "github.com/jackdoe/baxx/help"
	"github.com/marcusolsson/tui-go"
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
	p.SetBorder(false)
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
	ui.SetKeybinding("$", func() { scroll.ScrollToBottom() })
	ui.SetKeybinding("G", func() { scroll.ScrollToBottom() })
	ui.SetKeybinding("0", func() { scroll.ScrollToTop() })
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
	isRegisterMode := true
	form.AppendRow(tui.NewLabel("E-mail"))
	form.AppendRow(user)
	form.AppendRow(tui.NewLabel("Password"))
	form.AppendRow(password)
	form.AppendRow(tui.NewLabel("Confirm Password"))
	form.AppendRow(confirmPassword)

	register := tui.NewButton("[Register]")
	login := tui.NewButton("[Login]")

	quit := tui.NewButton("[Quit]")
	help := tui.NewButton("[Help]")
	pitch := tui.NewButton("[What/Why/How]")
	tos := tui.NewButton("[Terms Of Service]")
	buttonsRegister := tui.NewHBox(
		tui.NewSpacer(),
		tui.NewPadder(1, 0, register),
		tui.NewPadder(1, 0, login),
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
		tui.NewPadder(1, 0, tui.NewLabel(Render(HelpObject{Template: GuiInfo}))),
		tui.NewPadder(1, 0, tui.NewLabel("Contact Us:")),
		tui.NewPadder(1, 0, tui.NewLabel(" * Slack         https://baxx.dev/join/slack")),
		tui.NewPadder(1, 0, tui.NewLabel(" * Google Groups https://baxx.dev/join/groups")),

		tui.NewPadder(1, 1, form),
		tui.NewPadder(1, 0, tui.NewLabel("Registering means you agree with")),
		tui.NewPadder(1, 0, tui.NewLabel("the terms of service!")),
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
	chain.Set(user, password, confirmPassword, register, login, help, pitch, tos, quit)
	ui.SetFocusChain(chain)

	quit.OnActivated(func(b *tui.Button) {
		ui.Quit()
	})

	help.OnActivated(func(b *tui.Button) {
		popup(ui, root, false, nil, "HELP", Render(HelpObject{Template: AllHelp, Status: bcommon.EMPTY_STATUS}))
	})

	pitch.OnActivated(func(b *tui.Button) {
		popup(ui, root, false, nil, "HELP", Render(HelpObject{Template: GuiPitch}))
	})

	tos.OnActivated(func(b *tui.Button) {
		popup(ui, root, false, nil, "Terms Of Service", Render(HelpObject{Template: GuiTos}))
	})

	login.OnActivated(func(b *tui.Button) {
		if isRegisterMode {
			// copy pasta
			form.RemoveRows()
			form.AppendRow(tui.NewLabel("Email"))
			form.AppendRow(user)
			form.AppendRow(tui.NewSpacer())
			form.AppendRow(tui.NewLabel("Password"))
			form.AppendRow(password)
			form.AppendRow(tui.NewSpacer())
			user.SetFocused(true)
			isRegisterMode = false
			b.SetFocused(false)
			chain.Set(user, password, register, login, help, pitch, tos, quit)
			ui.SetFocusChain(chain)

			return
		}
		p1 := password.Text()
		email := user.Text()
		_, err := bc.Status(&bcommon.CreateUserInput{Email: email, Password: p1})
		if err == nil {
			onRegister(email, p1)
		} else {
			popup(ui, root, false, nil, "ERROR", apiError(err))
		}

	})

	register.OnActivated(func(b *tui.Button) {
		if !isRegisterMode {
			// copy pasta
			form.RemoveRows()
			form.AppendRow(tui.NewLabel("Email"))
			form.AppendRow(user)
			form.AppendRow(tui.NewSpacer())
			form.AppendRow(tui.NewLabel("Password"))
			form.AppendRow(password)
			form.AppendRow(tui.NewSpacer())
			form.AppendRow(tui.NewLabel("Confirm Password"))
			form.AppendRow(confirmPassword)
			isRegisterMode = true
			user.SetFocused(true)
			b.SetFocused(false)
			chain.Set(user, password, confirmPassword, register, login, help, pitch, tos, quit)
			ui.SetFocusChain(chain)
			return
		}

		p1 := password.Text()
		p2 := confirmPassword.Text()
		email := user.Text()
		if p1 != p2 {
			popup(ui, root, false, nil, "ERROR", "passwords must match")
			return
		}

		if p1 == "" {
			popup(ui, root, false, nil, "ERROR", Render(HelpObject{Template: GuiPassRequired}))
			return
		}
		if email == "" {
			popup(ui, root, false, nil, "ERROR", Render(HelpObject{Template: GuiEmailRequired}))
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
		popup(ui, content, false, nil, "HELP", Render(HelpObject{Template: GuiPassRequired}), "", Render(HelpObject{Template: AllHelp, Status: bcommon.EMPTY_STATUS}))
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
			subscribed.SetText("Activate at https://baxx.dev/sub/" + status.PaymentID + "\nIt takes 1-2 minutes after paying to enable the account.")
		}
		if status.Paid && status.EmailVerified != nil {
			stop <- true
			popup(ui, content, true, nil, "SUCCESS", "Your account is now ready to be used", "", "", Render(HelpObject{Template: AllHelp, Status: status}))
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
	server := os.Getenv("BAXX_REMOTE")
	if server == "" {
		server = "https://baxx.dev"
	}

	statusUpdate := make(chan string)

	bc := baxx.NewClient(nil, server, statusUpdate)
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
