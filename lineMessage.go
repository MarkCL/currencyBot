package main

import (
	"fmt"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"net/http"
	"os"
	"strings"
)

func lineCallbackHandler(w http.ResponseWriter, req *http.Request) {
	events, err := lineBotClient.ParseRequest(req)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				//fmt.Println(message.Text)
				go LineReplyTextMessage(event.ReplyToken, ProcessReplyMessage(event.Source.UserID , message.Text, false).(string))
			}
		} else if event.Type == linebot.EventTypeFollow {
			if _, ok := userTargetData[event.Source.UserID]; !ok {
				go LineGetApprovalForUserFollow(event.Source.UserID)
			}
		} else if event.Type == linebot.EventTypeUnfollow {
			if _, ok := userTargetData[event.Source.UserID]; ok {
				go LineRemoveUser(event.Source.UserID)
			}
		} else if event.Type == linebot.EventTypePostback {
			go LinePostBackApprovalResultForUserFollow(event.Postback.Data)
		}
	}
}

func LineReplyTextMessage(replyToken string, messageBody string) {
	ReplyMessage := linebot.NewTextMessage(messageBody)
	if _, err := lineBotClient.ReplyMessage(replyToken, ReplyMessage).Do(); err != nil {
		log.Print(err)
	}
}

func LineSendTextMessage(id string, messageBody string) {
	Message := linebot.NewTextMessage(messageBody)
	if _, err := lineBotClient.PushMessage(id, Message).Do(); err!= nil{
		log.Print(err)
	}
}

func LineGetApprovalForUserFollow(id string){
	profile, err := lineBotClient.GetProfile(id).Do()
	userInfo := "User Unique ID:" + id
	if err == nil{
		userInfo = "Name:" + profile.DisplayName
	}
	getApprovalMessage := linebot.NewButtonsTemplate("", "New User Add", userInfo, linebot.NewPostbackAction("Approve", "Approve", "Approve:" + id, "Approve:" + id), linebot.NewPostbackAction("Reject", "Reject", "Reject:" + id, "Reject:" + id))
	lineBotClient.PushMessage(os.Getenv("lineAdminID"), linebot.NewTemplateMessage("new user", getApprovalMessage))
}

func LinePostBackApprovalResultForUserFollow(dataFromPostBack string){
	newUserId := strings.Split(dataFromPostBack, ":")[1]
	if strings.Count(dataFromPostBack, "Reject") != 0 {
		if _, ok := userTargetData[newUserId]; ok{
			LineRemoveUser(newUserId)
		}
		lineBotClient.PushMessage(newUserId, linebot.NewTextMessage("未批准"))
	}else if strings.Count(dataFromPostBack, "Approve") != 0 {
		if _, ok := userTargetData[newUserId]; !ok {
			LineAddNewUser(newUserId)
		}
		lineBotClient.PushMessage(newUserId, linebot.NewTextMessage(WllComMessage))
	}
}

func LineAddNewUser(id string){
	userTargetData[id] = &userTarget{TargetCurrency: DefaultTargetCurrency, TargetPrice: DefaultTargetPrice, UsedApiTimes: 0}
	updateLineUserTargetsToDB(id, DefaultTargetCurrency, DefaultTargetPriceStr)
}

func LineRemoveUser(id string){
	delete(userTargetData, id)
	_, err := db.Exec("DELETE FROM user_profile WHERE user_id='" + id + "'")
	if err != nil{
		fmt.Println("remove line user from DB fail!")
	}
}