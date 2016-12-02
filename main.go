package main

import (
    "log"
    "net/http"
    "os"
    "github.com/gin-gonic/gin"
    "github.com/PuerkitoBio/goquery"
    "github.com/line/line-bot-sdk-go/linebot"
)

func ScrapeText(url string) string{
    res, err := goquery.NewDocument(url)
    if err != nil {
        log.Print(err)
    }
    title, _ := res.Find("h2").Html()
    text := make([]byte, 0, 10)
    text = append(text, title...)
    res.Find(".kcm-read-text > p").Each(func(_ int, s *goquery.Selection) {
        // text = append(text, s.Text()...)
    })
    return string(text)
}

func main() {
    port := os.Getenv("PORT")

    if port == "" {
        log.Fatal("$PORT must be set")
    }

    bot, err := linebot.New(
        os.Getenv("CHANNEL_SECRET"),
        os.Getenv("CHANNEL_TOKEN"),
    )
    if err != nil {
        log.Fatal(err)
    }

    r := gin.New()
    r.Use(gin.Logger())

    r.POST("/callback", func(c *gin.Context) {
        events, err := bot.ParseRequest(c.Request)
        if err != nil {
            if err == linebot.ErrInvalidSignature {
                log.Print(err)
            }
            return
        }
        for _, event := range events {
            if event.Type == linebot.EventTypeMessage {
                switch event.Message.(type) {
                case *linebot.TextMessage:
                    res, err := goquery.NewDocument("http://news.kompas.com/")
                    if err != nil {
                        log.Print(err)
                    }
                    url, _ := res.Find(".list-most > a").First().Attr("href")
                    news := ScrapeText(url)
                    if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(news)).Do(); err != nil {
                        log.Print(err)
                    }
                }
            }
        }
    })

    r.Run(":" + port)
}
