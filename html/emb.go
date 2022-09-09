package html

import _ "embed"

//go:embed captcha.html
var CaptchaHTML []byte

//go:embed result.html
var ResultHTML []byte
