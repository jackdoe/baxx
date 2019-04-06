package help

type TextTemplate int

//go:generate stringer -type=TextTemplate
const (
	EmailAfterRegistration TextTemplate = iota
	EmailNotification
	EmailPaymentCancel
	EmailPaymentPlease
	EmailValidation
	FileMeta
	SyncMeta
	GuiEmailRequired
	GuiInfo
	GuiPassRequired
	GuiPitch
	GuiTos
	HtmlLinkError
	HtmlLinkExpired
	HtmlVerificationOk
	HtmlWaitPaypal
	TokenMeta
	Profile
	AllHelp
	EmailQuotaLeft
)
