package models

type WPLoginRequestPayload struct {
	Log        string `form:"log"`
	PWD        string `form:"pwd"`
	WpSubmit   string `form:"wp-submit"`
	RedirectTo string `form:"redirect-to"`
	TestCookie string `form:"test-cookie"`
}

func NewWPLoginPayload(username, password, redirectURL string) *WPLoginRequestPayload {
	return &WPLoginRequestPayload{}
}
