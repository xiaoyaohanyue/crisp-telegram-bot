package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"crisp_tg_bot/utils"

	crisp "github.com/crisp-im/go-crisp-api/crisp/v3"
	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"
)

var bot *tgbotapi.BotAPI
var client *crisp.Client
var redisClient *redis.Client
var config *viper.Viper

// CrispMessageInfo stores the original message
type CrispMessageInfo struct {
	WebsiteID string
	SessionID string
}

// MarshalBinary serializes CrispMessageInfo into binary
func (s *CrispMessageInfo) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

// UnmarshalBinary deserializes CrispMessageInfo into struct
func (s *CrispMessageInfo) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func contains(s []interface{}, e int64) bool {
	for _, a := range s {
		if int64(a.(int)) == e {
			return true
		}
	}
	return false
}

func replyToUser(update *tgbotapi.Update) {
	if update.Message.ReplyToMessage == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "请回复一个消息")
		bot.Send(msg)
		return
	}

	res, err := redisClient.Get(strconv.Itoa(update.Message.ReplyToMessage.MessageID)).Result()

	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "ERROR: "+err.Error())
		bot.Send(msg)
		return
	}

	var msgInfo CrispMessageInfo
	err = json.Unmarshal([]byte(res), &msgInfo)

	if err := json.Unmarshal([]byte(res), &msgInfo); err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "ERROR: "+err.Error())
		bot.Send(msg)
		return
	}

	if update.Message.Text != "" {
		client.Website.SendTextMessageInConversation(msgInfo.WebsiteID, msgInfo.SessionID, crisp.ConversationTextMessageNew{
			Type:    "text",
			From:    "operator",
			Origin:  "chat",
			Content: update.Message.Text,
		})
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "回复成功！")
	bot.Send(msg)
}

func sendMsgToAdmins(text string, WebsiteID string, SessionID string) {
	for _, id := range config.Get("admins").([]interface{}) {
		msg := tgbotapi.NewMessage(int64(id.(int)), text)
		msg.ParseMode = "Markdown"
		sent, _ := bot.Send(msg)

		redisClient.Set(strconv.Itoa(sent.MessageID), &CrispMessageInfo{
			WebsiteID,
			SessionID,
		}, time.Duration(config.GetInt("redis.cacheTime"))*time.Hour)
	}
}
func sendMsgToAdminsHtml(text string, WebsiteID string, SessionID string) {
	for _, id := range config.Get("admins").([]interface{}) {
		msg := tgbotapi.NewMessage(int64(id.(int)), text)
		msg.ParseMode = "HTML"
		sent, _ := bot.Send(msg)

		redisClient.Set(strconv.Itoa(sent.MessageID), &CrispMessageInfo{
			WebsiteID,
			SessionID,
		}, time.Duration(config.GetInt("redis.cacheTime"))*time.Hour)
	}
}

func init() {
	config = utils.GetConfig()

	log.Printf("Initializing Redis...")

	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.GetString("redis.host"),
		Password: config.GetString("redis.password"),
		DB:       config.GetInt("redis.db"),
	})

	var err error

	_, err = redisClient.Ping().Result()
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Initializing Bot...")

	bot, err = tgbotapi.NewBotAPI(config.GetString("telegram.key"))

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = config.GetBool("debug")
	bot.RemoveWebhook()

	log.Printf("Authorized on account %s", bot.Self.UserName)

	log.Printf("Initializing Crisp Listner")
	client = crisp.New()
	// Set authentication parameters
	client.AuthenticateTier("plugin", config.GetString("crisp.identifier"), config.GetString("crisp.key"))

	client.Events.Listen(
		crisp.EventsModeWebSockets,

		[]string{
			"message:send",
			"message:received",
			"message:compose:send",
			"message:compose:receive",
		},

		func(reg *crisp.EventsRegister) {
			fmt.Print("WebSocket channel is connected: now listening for events\n")

			reg.On("message:send/text", func(evt crisp.EventsReceiveTextMessage) {
				fmt.Printf("[message:send/text] %s\n", evt)
				// text := fmt.Sprintf(`*%s(%s)发来消息: %%0A *%s`, *evt.User.Nickname, *evt.User.UserID, *evt.Content)
				text := fmt.Sprintf("用户(*%s*) 发来消息: \n\n %s", *evt.User.Nickname, *evt.Content)
				sendMsgToAdmins(text, *evt.WebsiteID, *evt.SessionID)
			})

			reg.On("message:send/file", func(evt crisp.EventsReceiveFileMessage) {
				fmt.Printf("[message:send/file] %s\n", evt)
				text := fmt.Sprintf(`用户(<b>%s</b>) 发来一张图片:%s)`, *evt.User.Nickname, evt.Content.URL)
				sendMsgToAdminsHtml(text, *evt.WebsiteID, *evt.SessionID)
			})

			reg.On("message:send/animation", func(evt crisp.EventsReceiveAnimationMessage) {
				fmt.Printf("[message:send/animation] %s\n", evt)
			})

			reg.On("message:send/audio", func(evt crisp.EventsReceiveAudioMessage) {
				fmt.Printf("[message:send/audio] %s\n", evt)
			})

			reg.On("message:send/picker", func(evt crisp.EventsReceivePickerMessage) {
				fmt.Printf("[message:send/picker] %s\n", evt)
			})

			reg.On("message:send/field", func(evt crisp.EventsReceiveFieldMessage) {
				fmt.Printf("[message:send/field] %s\n", evt)
			})

			reg.On("message:send/carousel", func(evt crisp.EventsReceiveCarouselMessage) {
				fmt.Printf("[message:send/carousel] %s\n", evt)
			})

			reg.On("message:send/note", func(evt crisp.EventsReceiveNoteMessage) {
				fmt.Printf("[message:send/note] %s\n", evt)
			})

			reg.On("message:send/event", func(evt crisp.EventsReceiveEventMessage) {
				fmt.Printf("[message:send/event] %s\n", evt)
			})

			reg.On("message:received/text", func(evt crisp.EventsReceiveTextMessage) {
				fmt.Printf("[message:received/text] %s\n", evt)
			})

			reg.On("message:received/file", func(evt crisp.EventsReceiveFileMessage) {
				fmt.Printf("[message:received/file] %s\n", evt)
			})

			reg.On("message:received/animation", func(evt crisp.EventsReceiveAnimationMessage) {
				fmt.Printf("[message:received/animation] %s\n", evt)
			})

			reg.On("message:received/audio", func(evt crisp.EventsReceiveAudioMessage) {
				fmt.Printf("[message:received/audio] %s\n", evt)
			})

			reg.On("message:received/picker", func(evt crisp.EventsReceivePickerMessage) {
				fmt.Printf("[message:received/picker] %s\n", evt)
			})

			reg.On("message:received/field", func(evt crisp.EventsReceiveFieldMessage) {
				fmt.Printf("[message:received/field] %s\n", evt)
			})

			reg.On("message:received/carousel", func(evt crisp.EventsReceiveCarouselMessage) {
				fmt.Printf("[message:received/carousel] %s\n", evt)
			})

			reg.On("message:received/note", func(evt crisp.EventsReceiveNoteMessage) {
				fmt.Printf("[message:received/note] %s\n", evt)
			})

			reg.On("message:received/event", func(evt crisp.EventsReceiveEventMessage) {
				fmt.Printf("[message:received/event] %s\n", evt)
			})

			reg.On("message:compose:send", func(evt crisp.EventsReceiveMessageComposeSend) {
				fmt.Printf("[message:compose:send] %s\n", evt)
			})

			reg.On("message:compose:receive", func(evt crisp.EventsReceiveMessageComposeReceive) {
				fmt.Printf("[message:compose:receive] %s\n", evt)
			})
		},

		func() {
			fmt.Print("WebSocket channel is disconnected: will try to reconnect\n")
		},

		func(err error) {
			fmt.Print("WebSocket channel error: may be broken\n")
		},
	)

}

func main() {
	var updates tgbotapi.UpdatesChannel

	log.Print("Start pooling")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ = bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("%s %s: %s", update.Message.From.FirstName, update.Message.From.LastName, update.Message.Text)

		switch update.Message.Command() {
		case "start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "极客Crisp客服助手")
			msg.ParseMode = "Markdown"
			bot.Send(msg)
		}

		if contains(config.Get("admins").([]interface{}), int64(update.Message.From.ID)) {
			replyToUser(&update)
		}
	}
}
