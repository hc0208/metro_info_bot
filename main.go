package main

import (
    "log"
    "os"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "time"
    "regexp"
    "github.com/gin-gonic/gin"
    "github.com/line/line-bot-sdk-go/linebot"
)

type TrainInfomation struct {
    Context                string    `json:"@context"`
    Id                     string    `json:"@id"`
    Type                   string    `json:"@type"`
    Date                   time.Time `json:"dc:date"`
    Valid                  time.Time `json:"dct:valid"`
    Operator               string    `json:"odpt:operator"`
    TimeOfOrigin           time.Time `json:"odpt:timeOfOrigin"`
    Railway                string    `json:"odpt:railway"`
    TrainInformationStatus string    `json:"odpt:trainInformationStatus"`
    TrainInformationText   string    `json:"odpt:trainInformationText"`
}

type TrainInformations []TrainInfomation

func FetchTrainInfo(message string) string{
    url := make([]byte, 0, 10)
    url = append(url, "https://api.tokyometroapp.jp/api/v2/datapoints?rdf:type=odpt:TrainInformation&acl:consumerKey="...)
    url = append(url, os.Getenv("CONSUMER_KEY")...)

    res, err := http.Get(string(url))
    if err != nil {
        log.Fatal(err)
    }

    defer res.Body.Close()

    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        log.Fatal(err)
    }

    var trains TrainInformations
    err = json.Unmarshal(body, &trains)
    if err != nil {
        log.Fatal(err)
    }

    for _, train := range trains {
        rep := regexp.MustCompile(`[A-Za-z]*odpt.Railway:TokyoMetro.`)
        text := make([]byte, 0, 10)
        text = rep.ReplaceAllString(train.Railway, ""...)
        text = append(text, train.TrainInformationText...)
        if len(train.TrainInformationStatus) > 0 {
            text = append(text, train.TrainInformationStatus...)
        }
        message += string(text)
    }
    return message
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
                switch message := event.Message.(type) {
                case *linebot.TextMessage:
                    text := FetchTrainInfo(message.Text)
                    if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(text)).Do(); err != nil {
                        log.Print(err)
                    }
                }
            }
        }
    })

    r.Run(":" + port)
}
