package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/XiaoMengXinX/Fish-Telegram-Captcha-bot/html"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang-jwt/jwt/v4"
)

var siteKey = os.Getenv("SITE_KEY")
var secretKey = os.Getenv("SECRET_KEY")

type VerifyResp struct {
	Success     bool      `json:"success"`
	ChallengeTs time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	Credit      bool      `json:"credit"`
}

func ChallengeHandler(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	var data JoinReqData
	bot := &tgbotapi.BotAPI{
		Token:  os.Getenv("BOT_TOKEN"),
		Client: &http.Client{},
		Buffer: 100,
	}
	bot.SetAPIEndpoint(tgbotapi.APIEndpoint)

	if token := r.Form.Get("token"); token == "" {
		t, _ := template.New("index").Parse(string(html.ResultHTML))
		_ = t.Execute(w, "Wrong parameters")
		return
	} else {
		var isVaild bool
		isVaild, data = VerifyJWT(token)
		if !isVaild {
			t, _ := template.New("index").Parse(string(html.ResultHTML))
			_ = t.Execute(w, "Incorrect parameters")
			return
		}
	}

	joinReqTime := time.Unix(data.Time, 0)
	if !joinReqTime.After(time.Now().Add(-180 * time.Second)) {
		t, _ := template.New("index").Parse(string(html.ResultHTML))
		_ = t.Execute(w, "Verification timeout, please resend your join request")
		return
	}

	if hCaptchaToken := r.Form.Get("g-recaptcha-response"); hCaptchaToken != "" {
		var resultText string
		t, _ := template.New("index").Parse(string(html.ResultHTML))
		result := VerifyCaptcha(hCaptchaToken)
		fmt.Println(result.Hostname, r.Host)
		switch {
		case !result.Success:
			resultText = "Verification failed, please close the page and try again"
		case !result.ChallengeTs.After(time.Now().Add(-60 * time.Second)):
			resultText = "Verification timeout, please close the page and try again"
		case result.Hostname != parseHostName(r.Host):
			resultText = "Verification failed, incorrect host name"
		default:
			_, _ = bot.Request(tgbotapi.ApproveChatJoinRequestConfig{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: data.ChatID,
				},
				UserID: data.UserID,
			})
			resultText = "Verification passed"
		}
		_ = t.Execute(w, resultText)
		return
	}

	t, _ := template.New("index").Parse(string(html.CaptchaHTML))
	_ = t.Execute(w, siteKey)
}

func VerifyJWT(tokenString string) (isVaild bool, data JoinReqData) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v ", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return false, data
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if str, ok := claims["data"].(string); ok {
			_ = json.Unmarshal([]byte(str), &data)
		} else {
			return false, data
		}
	} else {
		return false, data
	}
	return true, data
}

func VerifyCaptcha(token string) (r VerifyResp) {
	resp, _ := http.PostForm("https://hcaptcha.com/siteverify",
		url.Values{"secret": {secretKey}, "response": {token}},
	)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &r)
	return
}

func parseHostName(s string) string {
	domains := strings.Split(s, ".")
	if len(domains) < 2 {
		return ""
	}
	return strings.Join(domains[len(domains)-2:], ".")
}
