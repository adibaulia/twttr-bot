package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/labstack/echo/v4"
)

var (
	client              *twitter.Client
	CONSUMER_KEY_SECRET = os.Getenv("CONSUMER_SECRET_KEY")
	CONSUMER_KEY        = os.Getenv("CONSUMER_KEY")
	ACCESS_TOKEN        = os.Getenv("ACCESS_TOKEN")
	ACCESS_SECRET       = os.Getenv("ACCESS_SECRET")
)

type Request struct {
	Tweet     string `json:"tweet,omitempty"`
	Crc_token string `json:"crc_token,omitempty" form:"crc_token" query:"crc_token"`
}

type DM struct {
	DirectMessageEvent interface{} `json:"direct_message_event"`
}

func main() {
	port := os.Getenv("PORT")
	e := echo.New()
	tweet := e.Group("/tweet")
	tweet.POST("/create", createTweet)
	e.POST("/dev/webhooks", webhookEvent)
	e.GET("/dev/webhooks", CRC)
	e.Logger.Fatal(e.Start(":" + port))
}

func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}
func webhookEvent(c echo.Context) error {
	body := new(DM)
	if err := c.Bind(body); err != nil {
		log.Print("ERROR", err)
		return err
	}

	log.Print(body)
	return nil

}

func init() {
	config := oauth1.NewConfig(CONSUMER_KEY, CONSUMER_KEY_SECRET)
	token := oauth1.NewToken(ACCESS_TOKEN, ACCESS_SECRET)
	httpClient := config.Client(oauth1.NoContext, token)
	// Twitter client
	client = twitter.NewClient(httpClient)
	// for {
	// 	doEvery(2*time.Second, forFun)
	// }
}

func CRC(c echo.Context) error {
	body := new(Request)
	if err := c.Bind(body); err != nil {
		log.Print("ERROR", err)
		return err
	}

	secret := []byte(CONSUMER_KEY_SECRET)
	message := []byte(body.Crc_token)

	hash := hmac.New(sha256.New, secret)
	hash.Write(message)

	// to base64
	token := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	resp := map[string]string{
		"response_token": "sha256=" + token,
	}
	return c.JSON(http.StatusOK, resp)
}

func createTweet(c echo.Context) error {
	body := new(Request)
	if err := c.Bind(&body); err != nil {
		log.Print("ERROR", err)
		return err
	}
	// Send a Tweet
	tweet, resp, err := client.Statuses.Update(body.Tweet, nil)
	if err != nil {
		log.Print(err)
		return c.JSON(http.StatusBadRequest, http.Response{Status: err.Error()})
	}
	log.Print(resp)
	return c.JSON(http.StatusOK, http.Response{Status: tweet.Text})
}

func forFun(t time.Time) {
	now := time.Now().Format("030405")
	tweet := `_________ is the cutest girl i ever met ` + `(` + now + `)`
	_, resp, err := client.Statuses.Update(tweet, nil)
	if err != nil {
		log.Print(err)
	}
	log.Print(resp)
}
