package main

import (
	"fmt"
	baxx "github.com/jackdoe/baxx/client"
	bcommon "github.com/jackdoe/baxx/common"
	bhelp "github.com/jackdoe/baxx/help"
	"github.com/marcusolsson/tui-go"
	"log"
	"strings"
)

var logo = `
██████╗  █████╗ ██╗  ██╗██╗  ██╗
██╔══██╗██╔══██╗╚██╗██╔╝╚██╗██╔╝
██████╔╝███████║ ╚███╔╝  ╚███╔╝
██╔══██╗██╔══██║ ██╔██╗  ██╔██╗
██████╔╝██║  ██║██╔╝ ██╗██╔╝ ██╗
╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝
`

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
	buttons := tui.NewHBox(
		tui.NewSpacer(),
		tui.NewPadder(1, 0, help),
		tui.NewPadder(1, 0, register),
		tui.NewPadder(1, 0, quit),
	)

	window := tui.NewVBox(
		tui.NewPadder(10, 1, tui.NewLabel(logo)),
		tui.NewPadder(12, 0, tui.NewLabel("Welcome to baxx.dev!")),
		tui.NewPadder(1, 1, form),
		buttons,
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

	tui.DefaultFocusChain.Set(user, password, confirmPassword, help, register, quit)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

	popup := func(title string, buttonLabel string, msg ...string) {
		text := tui.NewVBox()

		for _, m := range msg {
			for _, s := range strings.Split(m, "\n") {
				text.Append(tui.NewLabel(s))
			}
		}

		scroll := tui.NewScrollArea(text)
		close := tui.NewButton(buttonLabel)
		close.SetFocused(true)
		p := tui.NewVBox(
			tui.NewPadder(1, 1, scroll),
			close,
		)

		p.SetBorder(true)
		close.SetSizePolicy(tui.Preferred, tui.Maximum)
		p.SetTitle(fmt.Sprintf("baxx.dev - %s", title))
		p.SetSizePolicy(tui.Expanding, tui.Minimum)
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

		ui.SetWidget(tui.NewPadder(5, 5, p))
	}

	quit.OnActivated(func(b *tui.Button) {
		ui.Quit()
	})
	help.OnActivated(func(b *tui.Button) {
		popup("SUCCESS", "[Back]", bhelp.GenericHelp())
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
