package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
)

func facebookMessagesEndPoint(w http.ResponseWriter, r *http.Request) {
	var callback Callback
	_ = json.NewDecoder(r.Body).Decode(&callback)
	if callback.Object == "page" {
		for _, entry := range callback.Entry {
			for _, event := range entry.Messaging {
				if !reflect.DeepEqual(event.Message, Message{}) {
					_, validUser := userTargetData[event.Sender.ID]
					if !invalidUserPool[event.Sender.ID] || event.Message.Text == TestTicket || validUser {
						go func() {
							if validUser {
								if userTargetData[event.Sender.ID].UsedApiTimes < (FacebookMaxApiTimes - 1) {
									FacebookSendTextMessage(ProcessReplyMessage(event.Sender.ID, event.Message.Text, true).(*bytes.Buffer))
								} else if userTargetData[event.Sender.ID].UsedApiTimes == (FacebookMaxApiTimes - 1) {
									FacebookSendTextMessage(ConvertStringToResponseBytes(event.Sender.ID, "超過單一小時最大訊息數(將於下一個小時恢復服務)\nExceed Max Time in this hour.(will recover service from begging of next hour.)"))
								}
								if userTargetData[event.Sender.ID].UsedApiTimes <= (FacebookMaxApiTimes - 1) {
									userTargetData[event.Sender.ID].UsedApiTimes += 1
									//fmt.Println("UsedApiTimes", userTargetData[event.Sender.ID].UsedApiTimes)
								}
							}else{
								FacebookSendTextMessage(ProcessReplyMessage(event.Sender.ID, event.Message.Text, true).(*bytes.Buffer))
							}
						}()
					}
				}
			}
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("Got your message"))
	} else {
		w.WriteHeader(404)
		_, _ = w.Write([]byte("Message not supported"))
	}
}

func facebookVerificationEndPoint(w http.ResponseWriter, r *http.Request) {
	challenge := r.URL.Query().Get("hub.challenge")
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	if mode != "" && token == os.Getenv("facebookMessengerPageToken") {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(challenge))
	} else {
		w.WriteHeader(404)
		_, _ = w.Write([]byte("Error, wrong validation token"))
	}
}

func FacebookSendTextMessage(body *bytes.Buffer) {
	client := http.Client{}
	url := fmt.Sprintf(FacebookMessengerApi, os.Getenv("facebookMessengerPageToken"))
	req, err := http.NewRequest("POST", url, body)
	if err == nil {
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
	} else {
		log.Fatal(err)
	}
}

type FacebookUserInfo struct {
	FirstName string  `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	ProfilePic string `json:"profile_pic,omitempty"`
	Locale string `json:"locale,omitempty"`
	Timezone int `json:"timezone,omitempty"`
	Gender string `json:"gender,omitempty"`
}

func FacebookGetUserName(id string) string {
	client := http.Client{}
	url := fmt.Sprintf(FacebookUserProfileApi, id, os.Getenv("facebookMessengerPageToken"))
	req, err := http.NewRequest("GET", url, nil)
	if err == nil {
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		var facebookUserInfo FacebookUserInfo
		_ = json.NewDecoder(resp.Body).Decode(&facebookUserInfo)
		return facebookUserInfo.LastName + facebookUserInfo.FirstName
	} else {
		log.Fatal(err)
	}
	return ""
}