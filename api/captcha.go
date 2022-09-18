package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/XiaoMengXinX/Fish-Telegram-Captcha-bot/html"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang-jwt/jwt/v4"
)

var siteKey = os.Getenv("SITE_KEY")
var secretKey = os.Getenv("SECRET_KEY")

type JoinReqData struct {
	UserID int64 `json:"user_id"`
	ChatID int64 `json:"chat_id"`
	Time   int64 `json:"time"`
	Type   int   `json:"type"`
}

type VerifyResp struct {
	Success     bool      `json:"success"`
	ChallengeTs time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	Credit      bool      `json:"credit"`
}

type queryData struct {
	Key   string
	Value string
}

type userData struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
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
		var isValid bool
		isValid, data = VerifyJWT(token)
		if !isValid {
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
		webappForm := r.Form.Get("webapp")
		result := VerifyCaptcha(hCaptchaToken)
		switch {
		case webappForm == "":
			resultText = "Invalid parameters, please open this page via telegram"
		case !VerifyWebappData(webappForm, data):
			resultText = "Incorrect parameters, this captcha is not for you"
		case !result.Success:
			resultText = "Verification failed, please close the page and try again"
		case !result.ChallengeTs.After(time.Now().Add(-60 * time.Second)):
			resultText = "Verification timeout, please close the page and try again"
		case result.Hostname != parseHostName(r.Host):
			resultText = "Verification failed, incorrect host name"
		default:
			if data.Type == 1 {
				chat, _ := bot.GetChat(tgbotapi.ChatInfoConfig{
					ChatConfig: tgbotapi.ChatConfig{
						ChatID: data.ChatID,
					},
				})
				action := tgbotapi.RestrictChatMemberConfig{
					ChatMemberConfig: tgbotapi.ChatMemberConfig{
						ChatID: data.ChatID,
						UserID: data.UserID,
					},
					Permissions: chat.Permissions,
				}
				_, _ = bot.Send(action)
			} else {
				_, _ = bot.Request(tgbotapi.ApproveChatJoinRequestConfig{
					ChatConfig: tgbotapi.ChatConfig{
						ChatID: data.ChatID,
					},
					UserID: data.UserID,
				})
			}
			resultText = "Verification passed"
		}
		_ = t.Execute(w, resultText)
		return
	}

	t, _ := template.New("index").Parse(string(html.CaptchaHTML))
	_ = t.Execute(w, siteKey)
}

func VerifyJWT(tokenString string) (isValid bool, data JoinReqData) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v ", token.Header["alg"])
		}
		return []byte(os.Getenv("BOT_TOKEN")), nil
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

func VerifyWebappData(webappData string, joinData JoinReqData) (isValid bool) {
	values, _ := url.ParseQuery(webappData)
	if len(values["hash"]) == 0 || len(values["user"]) == 0 {
		return false
	}
	dataCheckStr := parseDataCheckStr(values)
	secretKey := hmacSha256([]byte(os.Getenv("BOT_TOKEN")), []byte("WebAppData"))
	hash := hex.EncodeToString(hmacSha256([]byte(dataCheckStr), secretKey))
	if hash != values["hash"][0] {
		return false
	}
	var user userData
	_ = json.Unmarshal([]byte(values["user"][0]), &user)
	if user.ID != joinData.UserID {
		return false
	}
	return true
}

func hmacSha256(data, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(data)
	return h.Sum(nil)
}

func parseDataCheckStr(values url.Values) (dataCheckStr string) {
	var data []queryData
	for s, strs := range values {
		if s == "hash" {
			continue
		}
		data = append(data, queryData{
			Key:   s,
			Value: strs[0],
		})
	}
	sort.SliceStable(data, func(i, j int) bool {
		return data[i].Key[0] < data[j].Key[0]
	})
	for i, v := range data {
		dataCheckStr += fmt.Sprintf("%s=%s", v.Key, v.Value)
		if i != len(data)-1 {
			dataCheckStr += "\n"
		}
	}
	return
}

func parseHostName(s string) string {
	domains := strings.Split(s, ".")
	if len(domains) < 2 {
		return ""
	}
	return strings.Join(domains[len(domains)-2:], ".")
}
