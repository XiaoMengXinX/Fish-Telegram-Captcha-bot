package html

import _ "embed"

//go:embed hCaptcha.html
var HCaptchaHTML []byte

//go:embed reCaptcha.html
var ReCaptchaHTML []byte

//go:embed result.html
var ResultHTML []byte
