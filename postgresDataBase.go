package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func loadUserTargetFromDB() {
	rows, err := db.Query("SELECT * FROM user_profile")
	if err != nil {
		fmt.Println("Get data from DB Fail!")
	} else {
		for rows.Next() {
			var id, targetCurrency, targetPrice string
			err = rows.Scan(&id, &targetCurrency, &targetPrice)
			if err != nil {
				log.Fatalf("Scan error: %q\n", err)
			}
			tempTargetPrice, _ := strconv.ParseFloat(targetPrice,64)
			userTargetData[id] = &userTarget{StringTable[targetCurrency],tempTargetPrice, "", 0, false}
		}
	}

	rows, err = db.Query("SELECT * FROM facebookCurrencyBotUsersTarget")
	if err != nil {
		fmt.Println("Get data from DB Fail!")
	} else {
		for rows.Next() {
			var id, targetCurrency, targetPrice, Name string
			err = rows.Scan(&id, &targetCurrency, &targetPrice, &Name)
			if err != nil {
				log.Fatalf("Scan error: %q\n", err)
			}
			tempTargetPrice, _ := strconv.ParseFloat(targetPrice,64)
			userTargetData[id] = &userTarget{targetCurrency, tempTargetPrice, Name, 0, false}
		}
	}

	rows, err = db.Query("SELECT * FROM facebookCurrencyBotActiveKeys")
	if err != nil {
		fmt.Println("Get data from DB Fail!")
	} else {
		for rows.Next() {
			var activeCode, id string
			err = rows.Scan(&activeCode, &id)
			if err != nil {
				log.Fatalf("Scan error: %q\n", err)
			}
			userKeyIdMap[activeCode] = id
		}
	}

	if  _, ok := userTargetData[os.Getenv("lineAdminID")]; !ok {
		userTargetData[os.Getenv("lineAdminID")] = &userTarget{TargetCurrency:"美金 (USD)", TargetPrice: 30.2, UserName: "WuCheng-Lung", UsedApiTimes:0}
		go updateLineUserTargetsToDB(os.Getenv("lineAdminID"), strings.Split(userTargetData[os.Getenv("lineAdminID")].TargetCurrency," ")[0], fmt.Sprintf("%f",userTargetData[os.Getenv("lineAdminID")].TargetPrice))
	}
	if _, ok := userTargetData[os.Getenv("facebookAdminID")]; !ok {
		tempKey := preparedNewKey
		userKeyIdMap[tempKey] = os.Getenv("facebookAdminID")
		userTargetData[os.Getenv("facebookAdminID")] = &userTarget{TargetCurrency:"美金 (USD)", TargetPrice: 30.2, UserName: "WuCheng-Lung", UsedApiTimes:0}
		go updateFacebookUserTargetsToDB(os.Getenv("facebookAdminID"), userTargetData[os.Getenv("facebookAdminID")].TargetCurrency, fmt.Sprintf("%f",userTargetData[os.Getenv("facebookAdminID")].TargetPrice), userTargetData[os.Getenv("facebookAdminID")].UserName)
		go updateFacebookUserKeyIdMapToDB(tempKey, os.Getenv("facebookAdminID"))
		preparedNewKey = generateNewKey()
	}
}

func updateFacebookUserKeyIdMapToDB(activeCode string, id string) {
	if !userTargetData[id].TestUser {
		_, err := db.Exec("INSERT INTO facebookCurrencyBotActiveKeys (activeCode, userId) VALUES " +
			"('" + activeCode + "', '" + id + "') " +
			"ON CONFLICT(activeCode) DO NOTHING ")
		if err != nil {
			fmt.Println("Update DB Fail!")
		}
	}
}

func updateFacebookUserTargetsToDB(id string, targetCurrency string, targetPrice string, userName string) {
	if !userTargetData[id].TestUser {
		_, err := db.Exec("INSERT INTO facebookCurrencyBotUsersTarget (userId, userTargetCurrency, userTargetPrice, userName) VALUES " +
			"('" + id + "', '" + targetCurrency + "', '" + targetPrice + "', '" + userName + "') " +
			"ON CONFLICT(userId) DO UPDATE " +
			"SET userTargetCurrency='" + targetCurrency + "', userTargetPrice='" + targetPrice + "', userName='" + userName + "'")
		if err != nil {
			fmt.Println("Update DB Fail!")
		}
	}
}

func removeFacebookUserFromDB(Name string) error {
	_, err := db.Exec("DELETE FROM facebookCurrencyBotActiveKeys WHERE facebookCurrencyBotActiveKeys.userId=(SELECT userId FROM facebookCurrencyBotUsersTarget WHERE userName=" + "'" + Name + "'" + ")")
	if err != nil{
		fmt.Println("User ID is not found in activekey table!")
	}
	_, err = db.Exec("DELETE FROM facebookCurrencyBotUsersTarget WHERE userName=" + "'" + Name + "'")
	if err != nil{
		fmt.Println("User ID is not found in target table!")
	}
	return err
}

func updateLineUserTargetsToDB(id string, targetCurrency string, targetPrice string) {
	if !userTargetData[id].TestUser {
		rows, err := db.Query("SELECT * FROM user_profile WHERE user_id='" + id + "'")
		if err != nil {
			fmt.Println("Get data from DB Fail!")
		} else {
			if rows.Next() {
				_, err := db.Exec("UPDATE user_profile SET target_currency=" + "'" + targetCurrency + "', target_price='" + targetPrice + "' WHERE user_id='" + id + "'")
				if err != nil {
					fmt.Println("Update DB Fail!")
				}
			} else {
				_, err = db.Exec("INSERT INTO user_profile (user_id, target_currency, target_price) VALUES " + "('" + id + "', '" + targetCurrency + "', '"  + targetPrice + "');")
			}
		}
	}
}
