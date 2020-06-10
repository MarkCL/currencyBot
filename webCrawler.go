package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/context/ctxhttp"
	"log"
	"net/http"
	"strings"
	"time"
)

func DownloadCurrentPrice(url string, result *dataSet) {
	tr := &http.Transport{    //solve x509: certificate signed by unknown authority
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   60 * time.Second,
		Transport: tr,    //solve x509: certificate signed by unknown authority
	}

	req, err := http.NewRequest("GET", url, nil) //solve x509: certificate signed by unknown authority

	if err != nil {
		log.Println(err.Error())
		ReadFromBankWebSiteOK = false
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // solve request canceled
	defer cancel() // solve request canceled
	resp, err := ctxhttp.Do(ctx, client, req) // solve request canceled, 解決x509: certificate signed by unknown authority
	// resp, err := client.Do(req) //solve x509: certificate signed by unknown authority
	//resp, err := http.Get(url)//solve x509: certificate signed by unknown authority

	if err != nil {
		fmt.Println(err)
		ReadFromBankWebSiteOK = false
		return
	}

	defer func(){ _ = resp.Body.Close() }()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		fmt.Println(err)
		ReadFromBankWebSiteOK = false
		return
	}

	// Need to modify this part to obtain data for your target website >>>
	currencyInfo := doc.Find("body > div.page-wrapper > main > #ie11andabove > div > table > tbody > tr")
	if (len(currencyInfo.Nodes)) == 0{
		ReadFromBankWebSiteOK = false
		return
	}
	currencyInfo.Each(func(i int, s *goquery.Selection) {
		currency := strings.TrimSpace(s.Find("td.currency.phone-small-font > div > div.hidden-phone.print_show").Text())
		subResult := make(subset)
		subResult["Cash_Buy"] = s.Find("td.rate-content-cash.text-right.print_hide").Nodes[0].FirstChild.Data
		subResult["Cash_Sell"] = s.Find("td.rate-content-cash.text-right.print_hide").Nodes[1].FirstChild.Data
		subResult["NoneCash_Buy"] = s.Find("td.rate-content-sight.text-right.print_hide").Nodes[0].FirstChild.Data
		subResult["NoneCash_Sell"] = s.Find("td.rate-content-sight.text-right.print_hide").Nodes[1].FirstChild.Data
		(*result)[currency] = subResult
		//fmt.Printf("Review %d: currency:%s, Cash_Buy:%s, Cash_Sell:%s, NoneCash_Buy:%s, NoneCash_Sell:%s\n", i, currency, (*result)[currency]["Cash_Buy"], (*result)[currency]["Cash_Sell"], (*result)[currency]["NoneCash_Buy"], (*result)[currency]["NoneCash_Sell"])
		//fmt.Println("stringTable[\"" + strings.Split(currency, " ")[0] +"\"]" + "= \"" + currency + "\"")
		ReadFromBankWebSiteOK = true
	})
	// Need to modify this part to obtain data for your target website <<<
}
