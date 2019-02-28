package main

import (
	baxx "github.com/jackdoe/baxx/client"
	bcommon "github.com/jackdoe/baxx/common"
	"github.com/marcusolsson/tui-go"
	"log"
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
	bc := baxx.NewClient(nil, "http://localhost:8080", statusUpdate)

	user := tui.NewEntry()
	user.SetFocused(true)

	password := tui.NewEntry()
	password.SetEchoMode(tui.EchoModePassword)
	result := tui.NewVBox(tui.NewEntry())
	result.SetSizePolicy(tui.Preferred, tui.Maximum)
	confirmPassword := tui.NewEntry()
	confirmPassword.SetEchoMode(tui.EchoModePassword)

	form := tui.NewGrid(0, 0)
	form.AppendRow(tui.NewLabel("Email"))
	form.AppendRow(user)
	form.AppendRow(tui.NewSpacer())
	form.AppendRow(tui.NewLabel("Password"))
	form.AppendRow(password)
	form.AppendRow(tui.NewLabel("Confirm Password"))
	form.AppendRow(confirmPassword)
	status := tui.NewStatusBar("Ready.")
	go func() {
		for {
			s := <-statusUpdate
			status.SetText(s)
		}
	}()

	register := tui.NewButton("[Register]")

	quit := tui.NewButton("[Quit]")

	buttons := tui.NewHBox(
		tui.NewSpacer(),
		tui.NewPadder(1, 0, quit),
		tui.NewPadder(1, 0, register),
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
		result,
		status,
	)

	tui.DefaultFocusChain.Set(user, password, confirmPassword, register, quit)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

	quit.OnActivated(func(b *tui.Button) {
		ui.Quit()
	})

	register.OnActivated(func(b *tui.Button) {
		p1 := password.Text()
		p2 := confirmPassword.Text()
		email := user.Text()

		if p1 != p2 {
			status.SetText("--> passwords must match <--")
			return
		}
		if p1 == "" {
			status.SetText("--> password is required <--")
			return
		}
		if email == "" {
			status.SetText("--> email is required <--")
			return
		}

		u, err := bc.Register(&bcommon.CreateUserInput{Email: email, Password: p1})
		for i := 1; i < result.Length()+1; i++ {
			result.Remove(i)
		}

		if err != nil {
			result.SetTitle("ERROR")
			result.Append(tui.NewLabel(err.Error()))
			result.SetBorder(true)
		} else {
			result.SetTitle("SUCCESS")
			result.Append(tui.NewLabel("Semi Secret ID " + u.SemiSecretID + "\nMore information coming to your email\nThanks for using baxx.dev!"))
			result.SetBorder(true)
		}
	})

	ui.SetKeybinding("Esc", func() { ui.Quit() })

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}
