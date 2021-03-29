package notify

import (
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type NotifierTg struct {
	Token     string
	ChannelID int64
	ch        chan string
	bot       *tgbotapi.BotAPI
}

func NewNotifierTg(token string, channelID int64, httpClient *http.Client) *NotifierTg {
	// log.Printf("token: %s, channelID: %d", token, channelID)
	if token == "" || channelID == 0 {
		log.Printf("BotToken && ChannelID cannot be empty")
		return nil
	}
	bot, err := tgbotapi.NewBotAPIWithClient(token, httpClient)
	if err != nil {
		log.Printf("fail to init tgbot: %s", err)
		return nil
	}
	return &NotifierTg{
		Token:     token,
		ChannelID: channelID,
		ch:        make(chan string, 10),
		bot:       bot,
	}
}
func (n *NotifierTg) Start() {

	for _msg := range n.ch {
		go func(msg string) {
			tgMsg := tgbotapi.NewMessage(n.ChannelID, msg)
			tgMsg.DisableWebPagePreview = true
			_, _ = n.bot.Send(tgMsg)
		}(_msg)
	}

}
func (n *NotifierTg) Notify(msg string) {
	log.Printf("notifing %s", msg)
	n.ch <- msg
}
