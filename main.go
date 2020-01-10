package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/labstack/echo/v4"
)

var client *twitter.Client

type Request struct {
	Tweet string `json:"tweet,omitempty"`
}

func main() {
	port := os.Getenv("PORT")
	e := echo.New()
	tweet := e.Group("/tweet")
	tweet.POST("/create", createTweet)
	e.Logger.Fatal(e.Start(":" + port))
}

func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func init() {
	config := oauth1.NewConfig(os.Getenv("CONSUMER_KEY"), os.Getenv("CONSUMER_SECRET_KEY"))
	token := oauth1.NewToken(os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_SECRET"))
	httpClient := config.Client(oauth1.NoContext, token)
	// Twitter client
	client = twitter.NewClient(httpClient)
	for {
		doEvery(2*time.Second, forFun)
	}
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
	tweet := `NGELU NDASKU ` + `(` + now + `)`
	_, resp, err := client.Statuses.Update(tweet, nil)
	if err != nil {
		log.Print(err)
	}
	log.Print(resp)
}
