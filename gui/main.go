package main

import (
	"fmt"
	baxx "github.com/jackdoe/baxx/client"
	bcommon "github.com/jackdoe/baxx/common"
	bhelp "github.com/jackdoe/baxx/help"
	"github.com/marcusolsson/tui-go"
	"log"
)

var logo = `██████╗  █████╗ ██╗  ██╗██╗  ██╗
██╔══██╗██╔══██╗╚██╗██╔╝╚██╗██╔╝
██████╔╝███████║ ╚███╔╝  ╚███╔╝
██╔══██╗██╔══██║ ██╔██╗  ██╔██╗
██████╔╝██║  ██║██╔╝ ██╗██╔╝ ██╗
╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝`

/*
var logo = `
bbbbbbbb
b::::::b
b::::::b
b::::::b
 b:::::b
 b:::::bbbbbbbbb      aaaaaaaaaaaaa   xxxxxxx      xxxxxxxxxxxxxx      xxxxxxx
 b::::::::::::::bb    a::::::::::::a   x:::::x    x:::::x  x:::::x    x:::::x
 b::::::::::::::::b   aaaaaaaaa:::::a   x:::::x  x:::::x    x:::::x  x:::::x
 b:::::bbbbb:::::::b           a::::a    x:::::xx:::::x      x:::::xx:::::x
 b:::::b    b::::::b    aaaaaaa:::::a     x::::::::::x        x::::::::::x
 b:::::b     b:::::b  aa::::::::::::a      x::::::::x          x::::::::x
 b:::::b     b:::::b a::::aaaa::::::a      x::::::::x          x::::::::x
 b:::::b     b:::::ba::::a    a:::::a     x::::::::::x        x::::::::::x
 b:::::bbbbbb::::::ba::::a    a:::::a    x:::::xx:::::x      x:::::xx:::::x
 b::::::::::::::::b a:::::aaaa::::::a   x:::::x  x:::::x    x:::::x  x:::::x
 b:::::::::::::::b   a::::::::::aa:::a x:::::x    x:::::x  x:::::x    x:::::x
 bbbbbbbbbbbbbbbb     aaaaaaaaaa  aaaaxxxxxxx      xxxxxxxxxxxxxx      xxxxxxx
`
*/

func main() {
	statusUpdate := make(chan string)
	bc := baxx.NewClient(nil, "https://baxx.dev", statusUpdate)
	//	bc := baxx.NewClient(nil, "http://localhost:9123", statusUpdate)

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

	status := tui.NewStatusBar("")
	go func() {
		for {
			s := <-statusUpdate
			status.SetText(s)
		}
	}()

	register := tui.NewButton("[Register]")

	quit := tui.NewButton("[Quit]")
	help := tui.NewButton("[Help]")
	tos := tui.NewButton("[Terms Of Service]")
	buttonsTop := tui.NewHBox(
		tui.NewPadder(1, 0, help),
		tui.NewPadder(1, 0, tos),
		//		tui.NewSpacer(),
	)

	buttonsBottom := tui.NewHBox(
		tui.NewPadder(1, 0, register),
		tui.NewPadder(1, 0, quit),
		//		tui.NewSpacer(),
	)

	window := tui.NewVBox(
		tui.NewPadder(10, 1, tui.NewLabel(logo)),
		tui.NewPadder(1, 0, tui.NewLabel(bhelp.Intro())),
		tui.NewPadder(1, 1, form),
		tui.NewPadder(1, 0, tui.NewLabel("Registering means you agree with\nthe terms of service!")),
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

	root := tui.NewVBox(
		content,
		status,
	)

	tui.DefaultFocusChain.Set(user, password, confirmPassword, help, tos, register, quit)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

	popup := func(title string, buttonLabel string, msg ...string) {
		text := tui.NewVBox()

		scroll := tui.NewScrollArea(text)
		close := tui.NewButton(buttonLabel)
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
			if buttonLabel == "[Exit]" {
				ui.Quit()
			} else {
				ui.ClearKeybindings()
				ui.SetKeybinding("Esc", func() { ui.Quit() })
				ui.SetWidget(root)
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

		//text.Append(tui.NewLabel(fmt.Sprintf("%#v %#v ", text.SizeHint(), text.Size())))
		//text.Append(tui.NewLabel(fmt.Sprintf("%#v %#v", p.SizeHint(), p.Size())))
		//text.Append(tui.NewLabel(fmt.Sprintf("%#v %#v", scroll.SizeHint(), scroll.Size())))
		//text.Append(tui.NewLabel(fmt.Sprintf("%#v %#v", padder.SizeHint(), padder.Size())))

		for _, m := range msg {
			l := tui.NewLabel(m)
			l.SetSizePolicy(tui.Maximum, tui.Minimum)
			l.SetWordWrap(false)
			text.Append(l)
		}

	}

	quit.OnActivated(func(b *tui.Button) {
		ui.Quit()
	})
	help.OnActivated(func(b *tui.Button) {
		popup("HELP", "[Back]", bhelp.GenericHelp(), "", bhelp.AfterRegistration("WILL-BE-IN-YOUR-EMAIL", "your.email@example.com", "SECRET", "TOKEN-RW", "TOKEN-WO"))
	})

	tos.OnActivated(func(b *tui.Button) {
		popup("Terms Of Service", "[Back]", bhelp.TermsAndConditions())
	})

	register.OnActivated(func(b *tui.Button) {
		p1 := password.Text()
		p2 := confirmPassword.Text()
		email := user.Text()

		if p1 != p2 {
			popup("ERROR", "[Back]", "passwords must match")
			return
		}

		if p1 == "" {
			popup("ERROR", "[Back]", `
Password is required.

If you are not using a password manager
please use good passwords, such as: 

  'mickey mouse and metallica'

https://www.xkcd.com/936/`)
			return
		}
		if email == "" {
			popup("ERROR", "[Back]", `
Email is required.

It we will not send you any marketing messages,
it will be used just for business, such as:
 * sending notifications when backups are
   delayed, smaller than normal
 * payment is received
 * payment is not received
`)
			return
		}

		u, err := bc.Register(&bcommon.CreateUserInput{Email: email, Password: p1})
		if err != nil {
			popup("ERROR", "[Back]", fmt.Sprintf(`
API Error:

  %s

please contact help@baxx.dev if it persists`, err.Error()))
		} else {
			popup("SUCCESS", "[Exit]", u.Help)
		}
	})

	ui.SetKeybinding("Esc", func() { ui.Quit() })

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}
