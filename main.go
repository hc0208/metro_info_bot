package main

import (
    "log"
    "os"
    "fmt"
    "time"
    "regexp"
    "net/http"
    "io/ioutil"
    "encoding/json"
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

func fetchTrainName(railway string) string {
    name := map[string]string{
      "Ginza":      "銀座線",
      "Marunouchi": "丸の内線",
      "Chiyoda":    "千代田線",
      "Hibiya":     "日比谷線",
      "Namboku":    "南北線",
      "Yurakucho":  "有楽町線",
      "Fukutoshin": "副都心線",
      "Hanzomon":   "半蔵門線",
      "Tozai":      "東西線",
    }
    return name[railway]
}

func fetchTrainInfo(message string) string{
    info := "運行情報:\n"

    if message == "運行情報" {
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
            rep  := regexp.MustCompile(`[A-Za-z]*odpt.Railway:TokyoMetro.`)
            railway := rep.ReplaceAllString(train.Railway, "")
            railway  = fetchTrainName(railway)
            text := train.TrainInformationText
            if len(train.TrainInformationStatus) > 0 {
                text = fmt.Sprintf("%s (%s)", train.TrainInformationStatus, train.TrainInformationText)
            }
            info += fmt.Sprintf("%s: %s\n", railway, text)
        }
    } else {
        info = "「運行情報」と入力すると東京メトロの運行情報を表示します"
    }
    return info
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
                    text := fetchTrainInfo(message.Text)
                    if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(text)).Do(); err != nil {
                        log.Print(err)
                    }
                }
            }
        }
    })

    r.Run(":" + port)
}
