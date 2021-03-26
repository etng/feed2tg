package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/etng/feed2tg/opml"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mmcdole/gofeed"
)

func BuildOpml(title, ownerName, ownerEmail string) *opml.OPML {
	doc := &opml.OPML{
		Version: "2.0",
		Head: &opml.Head{
			Title:        title,
			DateCreated:  time.Now().Format(time.RFC1123),
			DateModified: time.Now().Format(time.RFC1123),
			OwnerName:    ownerName,
			OwnerEmail:   ownerEmail,
		},
		Body: &opml.Body{
			Outlines: []*opml.Outline{},
		},
	}
	return doc
}
func AddFeed(proxyURI string, doc *opml.OPML, feedURL, group string) {
	fp := gofeed.NewParser()
	fp.Client = ProxyClient(proxyURI)
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		panic(err)
	}
	var groupOutline *opml.Outline
	if group != "" {
		for _, outline := range doc.Body.Outlines {
			if outline.Text == group {
				groupOutline = outline
				break
			}
		}
		if groupOutline == nil {
			groupOutline = &opml.Outline{
				Text:     group,
				Outlines: []*opml.Outline{},
			}
			doc.Body.Outlines = append(doc.Body.Outlines, groupOutline)
		}
	}

	outline := &opml.Outline{
		Text:        feed.Title,
		Title:       feed.Title,
		Description: feed.Description,
		HTMLURL:     feed.Link,
		Language:    feed.Language,
		Type:        feed.FeedType,
		Version:     feed.FeedVersion,
		XMLURL:      feedURL,
	}
	if groupOutline != nil {
		groupOutline.Outlines = append(groupOutline.Outlines, outline)

	} else {
		doc.Body.Outlines = append(doc.Body.Outlines, outline)
	}
	log.Printf("added feed %s %s group %s", feed.Title, feedURL, group)
}
func main() {
	opts := InitOptions()
	notifyer := NewNotifierTg(opts.proxyURI, opts.TgToken, opts.TgChannelID)
	if notifyer != nil {
		notifyer.Start()
	}
	opmlFilename := "mine.opml"
	opmlDoc, e := opml.NewOPMLFromFile(opmlFilename)
	if e != nil {
		log.Printf("fail to read %s for %s, will genereate default one", opmlFilename, e)
		opmlDoc = BuildOpml("My OPML", "Yi Bo", "etng2004@gmail.com")
		for group, feedURLs := range map[string][]string{
			"Tech": {
				"http://feeds.feedburner.com/ruanyifeng",
				"https://me.ursb.me/feed",
				"https://winnielife.com/feed/",
				"https://blog.lilydjwg.me/feed",
			},
			"News": {
				"https://sspai.com/feed",
				"https://rsshub.app/sspai/matrix",
			},
		} {
			for _, feedURL := range feedURLs {
				AddFeed(opts.proxyURI, opmlDoc, feedURL, group)
			}
		}
		log.Printf("%#v", opmlDoc)
		of, _ := os.OpenFile(opmlFilename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
		defer of.Close()
		if xmlDoc, e := opmlDoc.XML(); e == nil {
			of.Write([]byte(xmlDoc))
		} else {
			log.Printf("can not generate opml %s", e)
		}
	}
	var lastUpdate time.Time
	lastUpdate = time.Now().UTC().Truncate(time.Hour * 24)
	log.Printf("last update is %s", lastUpdate)
	log.Printf("%#v", opmlDoc.Body)
	var UpdateNews = func() {
		log.Printf("start updating news now")
		startedAt := time.Now().UTC()
		log.Printf("%d outlines", len(opmlDoc.Body.Outlines))
		for _, outline := range opmlDoc.Body.Outlines {
			UpdateOutline("", lastUpdate, opts, notifyer, outline)
		}
		log.Printf("end updating news, used %s", time.Now().Sub(startedAt))
		lastUpdate = startedAt
	}
	UpdateNews()
	if opts.CheckInterval > 0 {
		log.Printf("will check in interval %s", opts.CheckInterval)
		ticker := time.NewTicker(opts.CheckInterval)
		defer ticker.Stop()
		for range ticker.C {
			UpdateNews()
		}
	}
	log.Printf("done")
}
func UpdateOutline(prefix string, lastUpdate time.Time, opts *Options, notifyer Notifyer, outline *opml.Outline) {
	if outline.XMLURL != "" {
		fp := gofeed.NewParser()
		fp.Client = ProxyClient(opts.proxyURI)
		feed, err := fp.ParseURL(outline.XMLURL)
		if err != nil {
			log.Printf("fail to parse outline %s %s for %s", outline.Title, outline.XMLURL, err)
		}
		for _, item := range feed.Items {
			if lastUpdate.IsZero() || item.PublishedParsed.UTC().After(lastUpdate) {
				author := ""
				if item.Author != nil {
					author = item.Author.Name
					if item.Author.Email != "" {
						author = fmt.Sprintf("%s <%s>", author, item.Author.Email)
					}
				}
				publishedAt := item.PublishedParsed.Format("2006-01-02 15:04")
				log.Printf("New Message: %s %s %s\n", publishedAt, author, item.Title)
				// fmt.Printf("%s\n", item.Description)
				// fmt.Printf("%s\n", item.Content)
				msg := strings.TrimSpace(fmt.Sprintf("%s %s %s %s %s", prefix, author, publishedAt, item.Title, item.Link))
				if notifyer != nil {
					notifyer.Notify(msg)
				}
			} else {
				// ignore old posts
			}
		}
	}
	for _, child := range outline.Outlines {
		UpdateOutline(outline.Text, lastUpdate, opts, notifyer, child)
	}
}

type Options struct {
	TgToken       string
	TgChannelID   int64
	proxyURI      string
	CheckInterval time.Duration
}

func InitOptions() (o *Options) {
	o = new(Options)
	flag.StringVar(&o.TgToken, "tg_token", "", "tt")
	flag.Int64Var(&o.TgChannelID, "tg_channel", 0, "tc")
	flag.StringVar(&o.proxyURI, "proxy", "", "")
	flag.DurationVar(&o.CheckInterval, "check_interval", time.Minute*0, "")
	flag.Parse()
	return o
}

type NotifierTg struct {
	Token     string
	ChannelID int64
	ch        chan string
	bot       *tgbotapi.BotAPI
}

func NewNotifierTg(proxyURI, token string, channelID int64) *NotifierTg {
	log.Printf("token: %s, channelID: %d", token, channelID)
	if token == "" || channelID == 0 {
		log.Printf("BotToken && ChannelID cannot be empty")
		return nil
	}
	bot, err := tgbotapi.NewBotAPIWithClient(token, ProxyClient(proxyURI))
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
	go func() {
		for msg := range n.ch {
			tgMsg := tgbotapi.NewMessage(n.ChannelID, msg)
			tgMsg.DisableWebPagePreview = true
			_, _ = n.bot.Send(tgMsg)
		}
	}()

}
func (n *NotifierTg) Notify(msg string) {
	log.Printf("notifing %s", msg)
	n.ch <- msg
}

func ProxyClient(proxyURI string) *http.Client {
	if proxyURI == "" {
		return &http.Client{}
	}
	var proxyURL *url.URL
	var e error
	if proxyURL, e = url.Parse(proxyURI); e != nil {
		log.Printf("bad proxyURI %s", proxyURI)
	}
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: time.Second * 60,
	}
}

type Notifyer interface {
	Notify(msg string)
}
