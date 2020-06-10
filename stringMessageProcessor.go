package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetCurrentPriceFromRAM(id string) string {
	return userTargetData[id].TargetCurrency +
		" 目標價(Target Price) " + strconv.FormatFloat(userTargetData[id].TargetPrice, 'f', -1, 64) +
		"\n現金賣出價(Cash Selling Price):\n" + obtainedData[userTargetData[id].TargetCurrency]["Cash_Sell"] +
		" NTD\n現金買入價(Cash Buying Price):\n" + obtainedData[userTargetData[id].TargetCurrency]["Cash_Buy"] +
		" NTD\n即期賣出價(Not-Cash Selling Price):\n" + obtainedData[userTargetData[id].TargetCurrency]["NoneCash_Sell"] +
		" NTD\n即期買入價(Not-Cash Buying Price):\n" + obtainedData[userTargetData[id].TargetCurrency]["NoneCash_Buy"] + " NTD"
}

func ConvertStringToResponseBytes(id string, inputText string) *bytes.Buffer {
	response := Response{
		Recipient: User{
			ID: id,
		},
		Message: Message{Text: inputText},
	}
	body := new(bytes.Buffer)
	_ = json.NewEncoder(body).Encode(response)
	return body
}

func ConvertCurrentPriceMessageToBytes(id string) *bytes.Buffer {
	body := ConvertStringToResponseBytes(id, GetCurrentPriceFromRAM(id))
	return body
}

var invalidUserPool = make(map[string]bool)
var userKeyIdMap = make (keyid)
var preparedNewKey = ""
var obtainedData = make(dataSet)
var userTargetData = make(map[string]*userTarget)
var minutesOfUpdateData uint64 = 6
var db *sql.DB
var dbUrl string
const TestTicket = "TEST_WUCHENGLUNGS_AWESOME_PROGRAM"

func startUpdateCurrencyData(ctx context.Context) {
	for{
		select{
		case <- ctx.Done():
			return
		default:
			time.Sleep(time.Minute * time.Duration(minutesOfUpdateData))
			location, _ := time.LoadLocation("Asia/Taipei")
			currentTime := time.Now().In(location)
			go func (){
				if currentTime.Weekday() != time.Sunday && currentTime.Weekday() != time.Saturday && currentTime.Hour() >= 8 && currentTime.Hour() < 17 {
					DownloadCurrentPrice(os.Getenv("BankURL"), &obtainedData)
					if ReadFromBankWebSiteOK{
						for id, target := range userTargetData {
							if curPrice, _ := strconv.ParseFloat(obtainedData[target.TargetCurrency]["Cash_Sell"], 64); curPrice <= target.TargetPrice {
								if len(id) == len(os.Getenv("lineAdminID")){
									LineSendTextMessage(id, GetCurrentPriceFromRAM(id))
								} else if len(id) == len(os.Getenv("facebookAdminID")) {
									if userTargetData[id].UsedApiTimes < (FacebookMaxApiTimes - 1){
										FacebookSendTextMessage(ConvertCurrentPriceMessageToBytes(id))
									} else if userTargetData[id].UsedApiTimes == (FacebookMaxApiTimes - 1){
										FacebookSendTextMessage(ConvertStringToResponseBytes(id, "超過單一小時最大訊息數(將於下一個小時恢復服務)\nExceed Max Time in this hour.(will recover service from begging of next hour.)"))
									}
									if userTargetData[id].UsedApiTimes <= (FacebookMaxApiTimes - 1){
										userTargetData[id].UsedApiTimes += 1
										//fmt.Println("UsedApiTimes", userTargetData[id].UsedApiTimes)
									}
								}
							}
						}
					}else{
						FacebookSendTextMessage(ConvertStringToResponseBytes(os.Getenv("facebookAdminID"), "Read from bank website failed."))
					}
				}
			}()
			//fmt.Println(currentTime.Hour())
			//fmt.Println(time.Now())
		}
	}
}

func ProcessReplyMessage(UserId string, TextMessageFromUser string, facebook bool) interface{} {
	_ , validUser := userTargetData[UserId]
	if validUser{
		switch{
		case TextMessageFromUser == "?":
			fallthrough
		case TextMessageFromUser == "？":
			DownloadCurrentPrice(os.Getenv("BankURL"), &obtainedData)
			if facebook {
				if ReadFromBankWebSiteOK {
					return ConvertCurrentPriceMessageToBytes(UserId)
				}
				return ConvertStringToResponseBytes(UserId, "Update Failed!")
			}else {
				if ReadFromBankWebSiteOK {
					return GetCurrentPriceFromRAM(UserId)
				}
				return "Update Failed!"
			}
		case strings.Contains(TextMessageFromUser, "目標價"):
			tempPrice := strings.Split(strings.ReplaceAll(TextMessageFromUser," ", ""), "目標價")[1]
			tempPriceFloat, err := strconv.ParseFloat(tempPrice, 64)
			if err != nil {
				if facebook{
					return ConvertStringToResponseBytes(UserId, "輸入資料或命令錯誤，請重新輸入!")
				}else{
					return "輸入資料或命令錯誤，請重新輸入!"
				}
			} else {
				userTargetData[UserId].TargetPrice = tempPriceFloat
				if facebook {
					go updateFacebookUserTargetsToDB(UserId, userTargetData[UserId].TargetCurrency, tempPrice, FacebookGetUserName(UserId))
					return ConvertStringToResponseBytes(UserId, "目標價已變更為: " + tempPrice)
				}else{
					go updateLineUserTargetsToDB(UserId, strings.Split(userTargetData[UserId].TargetCurrency, " ")[0], tempPrice)
					return "目標價已變更為: " + tempPrice
				}
			}
		case strings.Contains(strings.ToUpper(strings.ReplaceAll(TextMessageFromUser, " ", "")), "TARGETPRICE"):
			tempPrice := strings.Split(strings.ReplaceAll(strings.ToUpper(TextMessageFromUser)," ", "") , "TARGETPRICE")[1]
			tempPriceFloat, err := strconv.ParseFloat(tempPrice, 64)
			if err != nil {
				if facebook {
					return ConvertStringToResponseBytes(UserId, "Data format is or command error，please re-send it!")
				}else{
					return "Data format is or command error，please re-send it!"
				}
			} else {
				userTargetData[UserId].TargetPrice = tempPriceFloat
				if facebook{
					go updateFacebookUserTargetsToDB(UserId, userTargetData[UserId].TargetCurrency, tempPrice, FacebookGetUserName(UserId))
					return ConvertStringToResponseBytes(UserId, "Target Price has been changed to: " + tempPrice)
				}else{
					go updateLineUserTargetsToDB(UserId, strings.Split(userTargetData[UserId].TargetCurrency, " ")[0], tempPrice)
					return "Target Price has been changed to: " + tempPrice
				}
			}
		case  TextMessageFromUser == "MakeKey" && UserId == os.Getenv("facebookAdminID") && facebook:
			preparedNewKey = generateNewKey()
			return ConvertStringToResponseBytes(UserId, preparedNewKey)
		case TextMessageFromUser == "ActiveKey" && UserId == os.Getenv("facebookAdminID") && facebook:
			return ConvertStringToResponseBytes(UserId, preparedNewKey)
		case strings.Contains(TextMessageFromUser,"更新速率") && (UserId == os.Getenv("facebookAdminID") || UserId == os.Getenv("lineAdminID")):
			tempSpeed := strings.Split(TextMessageFromUser, "更新速率")[1]
			tempSpeedInt, err := strconv.ParseUint(tempSpeed, 10, 64)
			if err == nil{
				minutesOfUpdateData = tempSpeedInt
				if facebook{
					return ConvertStringToResponseBytes(UserId, "更新速率已變更為(data update duration changed): " + tempSpeed + "分鐘")
				}else{
					return "更新速率已變更為(data update duration changed): " + tempSpeed + "分鐘"
				}
			}
		case strings.Contains(TextMessageFromUser,"Speed") && (UserId == os.Getenv("facebookAdminID") || UserId == os.Getenv("lineAdminID")):
			tempSpeed := strings.Split(TextMessageFromUser, "Speed")[1]
			tempSpeedInt, err := strconv.ParseUint(tempSpeed, 10, 64)
			if err == nil{
				minutesOfUpdateData = tempSpeedInt
				if facebook{
					return ConvertStringToResponseBytes(UserId, "data update duration changed: " + tempSpeed + "mins")
				}else{
					return "data update duration changed: " + tempSpeed + "mins"
				}
			}
		case strings.Contains(TextMessageFromUser, "remove") && UserId == os.Getenv("facebookAdminID"):
			tempRemoveUserName := strings.ReplaceAll(strings.Split(TextMessageFromUser, "remove")[1], " ", "")
			err := removeFacebookUserFromDB(tempRemoveUserName)
			for k, v := range userTargetData{
				if v.UserName == tempRemoveUserName{
					delete(userTargetData, k)
				}
			}
			if err != nil{
				return ConvertStringToResponseBytes(UserId, "User '" + tempRemoveUserName + "' is not found in DB.")
			} else {
				return ConvertStringToResponseBytes(UserId, "User '" + tempRemoveUserName + "' has been removed.")
			}
		case TextMessageFromUser == "clear pool" && (UserId == os.Getenv("facebookAdminID") || UserId == os.Getenv("lineAdminID")):
			for k := range invalidUserPool{
				delete(invalidUserPool, k)
			}
			if facebook {
				return ConvertStringToResponseBytes(UserId, "Invalid user pool has been cleared.")
			} else {
				return 	"Invalid user pool has been cleared."
			}
		default:
			for key := range obtainedData {
				if strings.Contains(key, TextMessageFromUser) {
					userTargetData[UserId].TargetCurrency = key
					if facebook{
						return ConvertStringToResponseBytes(UserId,"目標幣別已變更為(Currency Target changed): " + key)
					}else{
						return "目標幣別已變更為(Currency Target changed): " + key
					}
				}
			}
		}
		if facebook{
			return ConvertStringToResponseBytes(UserId,"指令錯誤或查無此幣別!(Invalid command or not a currency type)")
		}else{
			return 	"指令錯誤或查無此幣別!(Invalid command or not a currency type)"
		}
	} else {
		if TextMessageFromUser == preparedNewKey && facebook{
			if _, ok := userKeyIdMap[preparedNewKey]; ok != false{
				return ConvertStringToResponseBytes(UserId, "這組啟用碼已被其他使用者啟用(This active code has been used by another user)!")
			} else {
				tempKey := preparedNewKey
				userKeyIdMap[preparedNewKey] = UserId
				userName := FacebookGetUserName(UserId)
				userTargetData[UserId] = &userTarget{TargetCurrency: DefaultTargetCurrency, TargetPrice: DefaultTargetPrice, UserName: userName, UsedApiTimes: 0}
				go updateFacebookUserTargetsToDB(UserId, DefaultTargetCurrency, DefaultTargetPriceStr, userName)
				go updateFacebookUserKeyIdMapToDB(tempKey, UserId)
				go func(){preparedNewKey = generateNewKey()}()
				go func(){
					prefix := "User "
					if userTargetData[UserId].TestUser{
						prefix = "Test " + prefix
					}
					FacebookSendTextMessage(ConvertStringToResponseBytes(os.Getenv("facebookAdminID"), prefix + userName + "' has been join this App."))
				}()
				return ConvertStringToResponseBytes(UserId, WllComMessage)
			}
		} else if !invalidUserPool[UserId] && strings.Contains(TextMessageFromUser,"ACTCB_") && facebook {
			invalidUserPool[UserId] = true
			tempKey := generateNewKey()
			for i := 0; i < 1; {
				if tempKey != preparedNewKey && tempKey != TextMessageFromUser{
					i = 1
				}
				tempKey = generateNewKey()
			}
			preparedNewKey = tempKey
			return ConvertStringToResponseBytes(UserId, "啟用碼錯誤(Invalid active code)!!")
		} else if TextMessageFromUser == TestTicket && facebook {
			userName := FacebookGetUserName(UserId)
			userTargetData[UserId] = &userTarget{TargetCurrency: DefaultTargetCurrency, TargetPrice: DefaultTargetPrice, UserName: userName, UsedApiTimes: 0, TestUser: true}
			go FacebookSendTextMessage(ConvertStringToResponseBytes(os.Getenv("facebookAdminID"), "User '" + userName + "' has been join this App."))
			return ConvertStringToResponseBytes(UserId, WllComMessage)
		} else if !invalidUserPool[UserId]{
			invalidUserPool[UserId] = true
			if facebook {
				return ConvertStringToResponseBytes(UserId, "非驗證過的使用者，需先取得管理者同意才能使用!\nInvalid user! To use this APP, you need to get approval from administartor first.")
			} else {
				return "非驗證過的使用者，需先取得管理者同意才能使用!\nInvalid user! To use this APP, you need to get approval from administartor first."
			}
		} else {
			return ""
		}
	}
}
