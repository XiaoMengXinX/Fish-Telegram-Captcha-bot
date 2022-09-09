package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = os.Getenv("SECRET_KEY")

type JoinReqData struct {
	UserID int64 `json:"user_id"`
	ChatID int64 `json:"chat_id"`
	Time   int64 `json:"time"`
}

func BotHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

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
		reqData := JoinReqData{
			UserID: update.ChatJoinRequest.From.ID,
			ChatID: update.ChatJoinRequest.Chat.ID,
			Time:   int64(update.ChatJoinRequest.Date),
		}
		reqDataJson, _ := json.Marshal(reqData)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"data": string(reqDataJson),
		})
		tokenString, _ := token.SignedString(jwtKey)
		msg := tgbotapi.NewMessage(update.ChatJoinRequest.Chat.ID, fmt.Sprintf("你正在申请加入群组「%s」，请点击下方按钮以完成加群验证。", update.ChatJoinRequest.Chat.Title))
		button := tgbotapi.NewInlineKeyboardButtonURL("开始验证", fmt.Sprintf("https://%s/captcha?token=%s", r.Host, tokenString))
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button))
		_, err = bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}
