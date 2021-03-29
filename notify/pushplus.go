package notify

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

// http://www.pushplus.plus/push2.html
const (
	// 支持html文本。为空默认使用html模板
	PushPlusTplHTML = "html"
	//  可视化展示json格式内容
	PushPlusTplJSON = "json"
	//  内容基于markdown格式展示
	PushPlusTplMarkdown = "markdown"
	//  阿里云监控报警定制模板
	PushPlusTplCloudMonitor = "cloudMonitor"
)

type NotifierPP struct {
	Token      string
	Topic      string
	ch         chan string
	httpClient *http.Client
}

func NewNotifierPP(token string, topic string, httpClient *http.Client) *NotifierPP {
	if token == "" {
		log.Printf("token cannot be empty")
		return nil
	}
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: time.Second * 30,
		}
	}
	return &NotifierPP{
		Token:      token,
		Topic:      topic,
		httpClient: httpClient,
		ch:         make(chan string, 10),
	}
}
func (n *NotifierPP) Start() {
	for _msg := range n.ch {
		go func(msg string) {
			body := map[string]interface{}{
				"token": n.Token,
				// "title":   msg,
				"content": msg,
			}
			if n.Topic != "" {
				body["topic"] = n.Topic
			}
			nb, _ := json.Marshal(body)
			req, _ := http.NewRequest(http.MethodPost, "http://www.pushplus.plus/send", bytes.NewBuffer(nb))
			req.Header.Set("Content-Type", "application/json")
			if resp, err := n.httpClient.Do(req); err != nil {
				log.Printf("fail to notify for %s", err)
			} else {
				rb, _ := httputil.DumpResponse(resp, true)
				log.Printf("notify response is %s", rb)
			}
		}(_msg)
	}
}
func (n *NotifierPP) Notify(msg string) {
	log.Printf("notifing %s", msg)
	n.ch <- msg
}
