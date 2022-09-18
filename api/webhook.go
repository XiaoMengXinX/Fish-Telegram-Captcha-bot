package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/XiaoMengXinX/Fish-Telegram-Captcha-bot/keywords"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang-jwt/jwt/v4"
)

var BlacklistKeywords = strings.Split(strings.ReplaceAll(string(keywords.Blacklist), "\r", ""), "\n")

func BotHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	if strings.ReplaceAll(r.URL.Path, "/webhook/", "") != os.Getenv("BOT_TOKEN") {
		_, _ = w.Write([]byte("bot token validation failed"))
		return
	}

	var update tgbotapi.Update

	err := json.Unmarshal(body, &update)
	if err != nil {
		return
	}

	bot := &tgbotapi.BotAPI{
		Token:  os.Getenv("BOT_TOKEN"),
		Client: &http.Client{},
		Buffer: 100,
	}
	bot.SetAPIEndpoint(tgbotapi.APIEndpoint)

	if update.ChatJoinRequest != nil {
		name := update.ChatJoinRequest.From.FirstName + " " + update.ChatJoinRequest.From.LastName
		if ContainsAny(name, BlacklistKeywords) || ContainsAny(update.ChatJoinRequest.Bio, BlacklistKeywords) {
			_, _ = bot.Request(tgbotapi.DeclineChatJoinRequest{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: update.ChatJoinRequest.Chat.ID,
				},
				UserID: update.ChatJoinRequest.From.ID,
			})
			return
		}
		reqData := JoinReqData{
			UserID: update.ChatJoinRequest.From.ID,
			ChatID: update.ChatJoinRequest.Chat.ID,
			Time:   int64(update.ChatJoinRequest.Date),
		}
		reqDataJson, _ := json.Marshal(reqData)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"data": string(reqDataJson),
		})
		tokenString, _ := token.SignedString([]byte(os.Getenv("BOT_TOKEN")))
		msg := tgbotapi.NewMessage(update.ChatJoinRequest.From.ID, fmt.Sprintf("你正在申请加入群组「%s」，请点击下方按钮以完成加群验证。\nYou are requesting to join the group 「%s」, please click the button below to complete the anti-spam verification.\n\n请在 180s 内完成加群验证\nPlease complete the verification within 180s.", update.ChatJoinRequest.Chat.Title, update.ChatJoinRequest.Chat.Title))
		webapp := tgbotapi.WebAppInfo{URL: fmt.Sprintf("https://%s/captcha?token=%s", r.Host, tokenString)}
		button := tgbotapi.InlineKeyboardButton{
			Text:   "开始验证/Verification",
			WebApp: &webapp,
		}
		//button := tgbotapi.NewInlineKeyboardButtonData("开始验证/Verification", fmt.Sprintf("https://%s/captcha?token=%s", r.Host, tokenString))
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button))
		_, _ = bot.Send(msg)
	}
}

func ContainsAny(str string, slice []string) bool {
	for _, v := range slice {
		if strings.Contains(str, v) {
			return true
		}
	}
	return false
}
