package help

type TextTemplate int

//go:generate stringer -type=TextTemplate
const (
	EmailAfterRegistration TextTemplate = iota
	EmailNotification
	EmailPaymentCancel
	EmailPaymentThanks
	EmailValidation
	FileDelete
	FileDownload
	FileList
	FileUpload
	FileMeta
	SyncShaOne
	SyncShaMany
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
	TokenCreate
	TokenDelete
	TokenList
	TokenModify
	TokenMeta
	Profile
	NotificationMeta
)
