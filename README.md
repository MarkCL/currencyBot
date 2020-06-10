# currencyBot
Combined Web crawler and FacebookMessenger/Line API in Golang to notify currency rate with message. This software can be deploy/run on heroku.

To use this software:
1. Apply a facebook developer account and a line developer account.
2. Create a web server or heroku application url for webhook in developer account in facebook and line.(add "/lineCallback" in the tail for line's webhook setting, "/callback" in the tail for facebook messenger's webhook setting)
3. Prepare a postgres database for this software.
4. Define some environment variables (If you are using heroku, you can set these variables on the heroku's website):  
    BankURL: this is the website url you want to track.  
    DATABASE_URL: this is the postgres database url you want to use.  
    facebookAdminID: Can be found in developer account.  
    facebookMessengerPageToken: Can be found in developer account.  
    lineAdminID: Can be found in developer account.  
    lineChannelSecret: Can be found in developer account.  
    lineChannelToken: Can be found in developer account.  
    PORT: default is 3000 in this program.  
5. Modify the part of web crawler in webCrawler.go for your target website to obtain correct data from web page.  
  
After everything is setted up, you can send message to this bot and it will reply.  
There are some commands:  
1. Send ? to the bot. --> This bot will reply current price.  
2. Send Target Price + "price" to the bot. --> for ex: Target Price 30.2 will set target price to 30.2, bot will inform you once target currency price meet you target.  
3. Send "currency" to the bot. --> for ex: USD will set Target currency to USD, and JPY will set Target currency to JPY.  
  
PS: This software will check the price from target website every 5 minutes.  
  
For security, I will not provide any environment variable I used in this software.  
