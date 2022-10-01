package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/XiaoMengXinX/Fish-Telegram-Captcha-bot/api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang-jwt/jwt/v4"
)

var botToken = os.Getenv("BOT_TOKEN")
var botAPI = os.Getenv("BOT_API")
var frontEndURL = os.Getenv("FRONTEND_URL")

func main() {
	apiEndpoint := tgbotapi.APIEndpoint
	if botAPI != "" {
		apiEndpoint = botAPI
	}
	bot, err := tgbotapi.NewBotAPIWithAPIEndpoint(botToken, apiEndpoint)
	if err != nil {
		log.Fatalln(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	u.AllowedUpdates = []string{"message", "chat_member", "chat_join_request"}

	updates := bot.GetUpdatesChan(u)
	defer bot.StopReceivingUpdates()

	for update := range updates {
		var reqData api.JoinReqData
		var chatID int64
		var chatTitle string

		if update.Message != nil {
			if update.Message.IsCommand() && update.Message.Command() == "start" && update.Message.Chat.IsPrivate() {
				if update.Message.CommandArguments() == "" {
					continue
				}
				groupID, _ := strconv.Atoi(update.Message.CommandArguments())
				member, _ := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
					ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
						ChatID: int64(groupID),
						UserID: update.Message.From.ID,
					},
				})
				if member.Status != "restricted" || member.CanSendMessages {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "你在目标群组中无需验证。\nThere is no need for you to verify in the target group."))
					continue
				}
				reqData.ChatID = int64(groupID)
				reqData.UserID = update.Message.From.ID
				reqData.Time = time.Now().Unix()
				reqData.Type = 1
				chatID = update.Message.Chat.ID
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("请点击下方按钮以完成加群验证。\nPlease click the button below to complete the anti-spam verification.\n\n请在 120s 内完成加群验证\nPlease complete the verification within 120s."))
				webapp := tgbotapi.WebAppInfo{URL: fmt.Sprintf("%s/captcha?token=%s", frontEndURL, getReqToken(reqData))}
				button := tgbotapi.InlineKeyboardButton{
					Text:   "开始验证/Verification",
					WebApp: &webapp,
				}
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button))
				_, err = bot.Send(msg)
				if err != nil {
					log.Printf("Send message to %d error: %v", chatID, err)
				}
				continue
			}
			if len(update.Message.NewChatMembers) > 0 {
				myRights, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
					ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
						ChatID: update.Message.Chat.ID,
						UserID: bot.Self.ID,
					},
				})
				if err != nil {
					log.Printf("GetMyRights on chat %d error: %v", update.Message.Chat.ID, err)
				}
				if !myRights.CanRestrictMembers || myRights.CanInviteUsers {
					continue
				}
				if update.Message.NewChatMembers[0].IsBot {
					continue
				}

				name := update.Message.NewChatMembers[0].FirstName + " " + update.Message.NewChatMembers[0].LastName
				if api.ContainsAny(name, api.BlacklistKeywords) || api.ContainsAny(update.Message.NewChatMembers[0].UserName, api.BlacklistKeywords) {
					action := tgbotapi.BanChatMemberConfig{
						ChatMemberConfig: tgbotapi.ChatMemberConfig{
							ChatID: update.Message.Chat.ID,
							UserID: update.Message.NewChatMembers[0].ID,
						},
					}
					_, _ = bot.Send(action)
					return
				}

				permissions := tgbotapi.ChatPermissions{
					CanSendMessages:       false,
					CanSendMediaMessages:  false,
					CanSendPolls:          false,
					CanSendOtherMessages:  false,
					CanAddWebPagePreviews: false,
					CanChangeInfo:         false,
					CanInviteUsers:        false,
					CanPinMessages:        false,
				}
				restrict := tgbotapi.RestrictChatMemberConfig{
					ChatMemberConfig: tgbotapi.ChatMemberConfig{
						ChatID: update.Message.Chat.ID,
						UserID: update.Message.NewChatMembers[0].ID,
					},
					Permissions: &permissions,
				}
				_, _ = bot.Send(restrict)

				reqData = api.JoinReqData{
					UserID: update.Message.NewChatMembers[0].ID,
					ChatID: update.Message.Chat.ID,
					Time:   int64(update.Message.Date),
					Type:   1,
				}
				chatID = update.Message.Chat.ID
				chatTitle = update.Message.Chat.Title
			}
		}

		if update.ChatJoinRequest != nil {
			name := update.ChatJoinRequest.From.FirstName + " " + update.ChatJoinRequest.From.LastName
			if api.ContainsAny(name, api.BlacklistKeywords) || api.ContainsAny(update.ChatJoinRequest.Bio, api.BlacklistKeywords) {
				_, _ = bot.Request(tgbotapi.DeclineChatJoinRequest{
					ChatConfig: tgbotapi.ChatConfig{
						ChatID: update.ChatJoinRequest.Chat.ID,
					},
					UserID: update.ChatJoinRequest.From.ID,
				})
				return
			}
			reqData = api.JoinReqData{
				UserID: update.ChatJoinRequest.From.ID,
				ChatID: update.ChatJoinRequest.Chat.ID,
				Time:   int64(update.ChatJoinRequest.Date),
			}
			chatID = update.ChatJoinRequest.From.ID
			chatTitle = update.ChatJoinRequest.Chat.Title
		}

		if reqData.Type == 1 {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("新成员你好，请点击下方按钮以完成加群验证。\nPlease click the button below to complete the anti-spam verification.\n\n请在 120s 内完成加群验证\nPlease complete the verification within 120s."))
			url := fmt.Sprintf("https://t.me/%s?start=%d", bot.Self.UserName, reqData.ChatID)
			button := tgbotapi.InlineKeyboardButton{
				Text: "开始验证/Verification",
				URL:  &url,
			}
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button))
			msg.ReplyToMessageID = update.Message.MessageID
			newMsg, err := bot.Send(msg)
			if err != nil {
				log.Printf("Send message to %d error: %v", chatID, err)
			}
			go CheckUserChatJoinStatus(bot, []int{newMsg.MessageID, update.Message.MessageID}, reqData)
		} else {
			if chatID == 0 {
				continue
			}
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("你正在申请加入群组「%s」，请点击下方按钮以完成加群验证。\nYou are requesting to join the group 「%s」, please click the button below to complete the anti-spam verification.\n\n请在 180s 内完成加群验证\nPlease complete the verification within 180s.", chatTitle, chatTitle))
			webapp := tgbotapi.WebAppInfo{URL: fmt.Sprintf("%s/captcha?token=%s", frontEndURL, getReqToken(reqData))}
			button := tgbotapi.InlineKeyboardButton{
				Text:   "开始验证/Verification",
				WebApp: &webapp,
			}
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button))
			_, err := bot.Send(msg)
			if err != nil {
				log.Printf("Send message to %d error: %v", chatID, err)
			}
		}
	}
}

func CheckUserChatJoinStatus(bot *tgbotapi.BotAPI, msgID []int, req api.JoinReqData) {
	userID := req.UserID
	chatID := req.ChatID
	timeStart := time.Now().Unix()
	for {
		userPermission, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: chatID,
				UserID: userID,
			},
		})
		if err != nil {
			log.Printf("Get user %d permission on chat %d error: %v", userID, chatID, err)
		}
		if userPermission.CanSendMessages {
			for _, id := range msgID {
				action := tgbotapi.NewDeleteMessage(chatID, id)
				_, _ = bot.Send(action)
			}
			break
		}
		if time.Now().Unix()-timeStart > 125 {
			ban := tgbotapi.BanChatMemberConfig{
				ChatMemberConfig: tgbotapi.ChatMemberConfig{
					ChatID: chatID,
					UserID: userID,
				},
				UntilDate: time.Now().Unix() + 40,
			}
			_, _ = bot.Send(ban)
			for _, id := range msgID {
				action := tgbotapi.NewDeleteMessage(chatID, id)
				_, _ = bot.Send(action)
			}
			break
		}
		time.Sleep(20 * time.Second)
	}
}

func getReqToken(data api.JoinReqData) string {
	reqDataJson, _ := json.Marshal(data)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data": string(reqDataJson),
	})
	tokenString, _ := token.SignedString([]byte(botToken))
	return tokenString
}
