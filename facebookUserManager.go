package main

import (
	"context"
	"math/rand"
	"os"
	"time"
)

func generateNewKey() string {
	var letterRunes = []rune("abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNOPQRSTUVWXYZ123456789")
	b := make([]rune, 20)
	for {
		for i := range b {
			b[i] = letterRunes[rand.Intn(len(letterRunes))]
		}
		if _, ok := userKeyIdMap["ACTCB_" + string(b)]; !ok {
			break
		}
	}
	return "ACTCB_" + string(b)
}

func startSweepTestUsers(ctx context.Context){
	for{
		select{
		case <- ctx.Done():
			return
		default:
			for id, _ := range userTargetData {
				if len(id) == len(os.Getenv("facebookAdminID")) && userTargetData[id].TestUser {
					delete(userTargetData, id)
				}
			}
			time.Sleep(time.Minute* time.Duration(120))
		}
	}
}

func startSweepFacebookApiTimesCounter(ctx context.Context) {
	for {
		select{
		case <- ctx.Done():
			return
		default:
			if time.Now().Minute() == 0 && time.Now().Second() == 0 {
				for _, v := range userTargetData{
					v.UsedApiTimes = 0
				}
				time.Sleep(time.Minute* time.Duration(1))
			}
		}
	}
}

func startUpdatePreparedKeyAndClearInvalidUserPool(ctx context.Context) {
	for {
		select{
		case <- ctx.Done():
			return
		default:
			time.Sleep(time.Minute * time.Duration(5))
			preparedNewKey = generateNewKey()
			for k := range invalidUserPool{
				delete(invalidUserPool, k)
			}
		}
	}
}
