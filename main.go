package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type userTarget struct{
	TargetCurrency string
	TargetPrice float64
	UserName string
	UsedApiTimes int64
	TestUser	bool
}
type keyid map[string]string
type dataSet map[string]map[string]string
type subset map[string]string

const(
	DefaultTargetCurrency = "日圓 (JPY)"
	DefaultTargetPrice    = 0.3
	DefaultTargetPriceStr = "0.3"
	FacebookMessengerApi  = "https://graph.facebook.com/v4.0/me/messages?access_token=%s"
	FacebookUserProfileApi = "https://graph.facebook.com/%s?fields=first_name,last_name,profile_pic&access_token=%s"
	FacebookMaxApiTimes = 200
	WllComMessage = "Welcome to Use this APP!\n\n" + "使用說明:\n" + "'?' --> 回應目前銀行價格\n" + "幣別 --> 設定目標幣別(舉例: 日圓)\n" + "目標價 + '價格' --> 設定現金賣出目標價(舉例: 目標價0.278)\n\n"+
		"Instructions for use:\n" + "'?' --> reply current price from Bank Website.\n" + "Currency Target Name --> Set Target Currency(ex: JPY).\n" + "Target + 'Price' --> Set Target Cash Sell Price(ex: TargetPrice0.278).\n\n"
)

var ReadFromBankWebSiteOK = false
var StringTable = make(map[string]string)
var lineBotClient *linebot.Client

func InitFullCurrencyNameFromChinese() {
	StringTable["美金"]= "美金 (USD)"
	StringTable["港幣"]= "港幣 (HKD)"
	StringTable["英鎊"]= "英鎊 (GBP)"
	StringTable["澳幣"]= "澳幣 (AUD)"
	StringTable["加拿大幣"]= "加拿大幣 (CAD)"
	StringTable["新加坡幣"]= "新加坡幣 (SGD)"
	StringTable["瑞士法郎"]= "瑞士法郎 (CHF)"
	StringTable["日圓"]= "日圓 (JPY)"
	StringTable["南非幣"]= "南非幣 (ZAR)"
	StringTable["瑞典幣"]= "瑞典幣 (SEK)"
	StringTable["紐元"]= "紐元 (NZD)"
	StringTable["泰幣"]= "泰幣 (THB)"
	StringTable["菲國比索"]= "菲國比索 (PHP)"
	StringTable["印尼幣"]= "印尼幣 (IDR)"
	StringTable["歐元"]= "歐元 (EUR)"
	StringTable["韓元"]= "韓元 (KRW)"
	StringTable["越南盾"]= "越南盾 (VND)"
	StringTable["馬來幣"]= "馬來幣 (MYR)"
	StringTable["人民幣"]= "人民幣 (CNY)"
}

func rootHandler(w http.ResponseWriter, req *http.Request){
	_, _ = w.Write([]byte("Hello this is a test page!"))
}

func handleLineAndFacebookMessages() {
	var err error
	lineBotClient, err = linebot.New(
		os.Getenv("lineChannelSecret"),
		os.Getenv("lineChannelToken"),
	)
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/lineCallback", lineCallbackHandler).Methods("POST")
	r.HandleFunc("/callback", facebookVerificationEndPoint).Methods("GET")
	r.HandleFunc("/callback", facebookMessagesEndPoint).Methods("POST")
	r.HandleFunc("/", rootHandler).Methods("GET")
	err = http.ListenAndServe(":" + os.Getenv("PORT"), r)
	if err != nil{
		log.Fatal(err)
	}
}

func Initialize(ctx context.Context) {
	InitFullCurrencyNameFromChinese()
	dbUrl = os.Getenv("DATABASE_URL")
	DownloadCurrentPrice(os.Getenv("BankURL"), &obtainedData)
	rand.Seed(time.Now().UnixNano())
	preparedNewKey = generateNewKey()
	var err error
	db, err = sql.Open("postgres", dbUrl)
	if err != nil{
		fmt.Println("Link SQL Sever Failed!")
	}else{
		err = db.Ping()
		if err != nil {
			panic(err)
		}
		fmt.Println("Postgres DataBase Successfully connected!")

		//_, err = db.Exec("DROP TABLE user_profile;")
		//if err != nil{
		//	fmt.Println("drop user_profile Table Failed!")
		//}
		//_, err = db.Exec("DROP TABLE facebookCurrencyBotUsersTarget;")
		//if err != nil{
		//	fmt.Println("drop facebookCurrencyBotUsersTarget Table Failed!")
		//}
		//_, err = db.Exec("DROP TABLE facebookCurrencyBotActiveKeys;")
		//if err != nil{
		//	fmt.Println("drop facebookCurrencyBotActiveKeys Table Failed!")
		//}
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS user_profile(user_id TEXT, target_currency TEXT, target_price FLOAT);") //Line's Database
		if err != nil{
			fmt.Println("create user_profile Table Failed!")
		}
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS facebookCurrencyBotUsersTarget(userId TEXT, userTargetCurrency TEXT, userTargetPrice FLOAT, userName TEXT, UNIQUE(userId));")
		if err != nil{
			fmt.Println("create facebookCurrencyBotUsersTarget Table Failed!")
		}
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS facebookCurrencyBotActiveKeys(activeCode TEXT, userId TEXT, UNIQUE(activeCode));")
		if err != nil{
			fmt.Println("create facebookCurrencyBotActiveKeys Table Failed!")
		}

		fmt.Println("Finished creat table (if not existed)")
	}

	loadUserTargetFromDB()
	go startUpdateCurrencyData(ctx)
	go startUpdatePreparedKeyAndClearInvalidUserPool(ctx)
	go startSweepFacebookApiTimesCounter(ctx)
	go startSweepTestUsers(ctx)
}



func main() {
	ctx, cancel := context.WithCancel(context.Background())
	Initialize(ctx)
	defer func(){_ = db.Close()}()
	handleLineAndFacebookMessages()
	cancel()
}
