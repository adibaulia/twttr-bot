package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/labstack/echo/v4"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"

	"google.golang.org/api/option"
)

var (
	client              *twitter.Client
	Conn                Connection
	CONSUMER_KEY_SECRET = os.Getenv("CONSUMER_SECRET_KEY")
	CONSUMER_KEY        = os.Getenv("CONSUMER_KEY")
	ACCESS_TOKEN        = os.Getenv("ACCESS_TOKEN")
	ACCESS_SECRET       = os.Getenv("ACCESS_SECRET")
)

type Request struct {
	Tweet     string `json:"tweet,omitempty"`
	Crc_token string `json:"crc_token,omitempty" form:"crc_token" query:"crc_token"`
}

type DMEvent struct {
	ForUserID           string                       `json:"for_user_id"`
	DirectMessageEvents []twitter.DirectMessageEvent `json:"direct_message_events"`
}

type Connection struct {
	DBConn *db.Client
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
	body := new(DMEvent)
	if err := c.Bind(body); err != nil {
		log.Print("ERROR", err)
		return err
	}
	ctx := context.Background()

	for _, val := range body.DirectMessageEvents {
		log.Print(val.Message.Data.Text)
		if (strings.Contains(val.Message.Data.Text, "HI!") || strings.Contains(val.Message.Data.Text, "hi!") || strings.Contains(val.Message.Data.Text, "Hi!")) && val.Message.SenderID != "1215181869567725568" {

			// if err := Conn.DBConn.NewRef("messages").Set(ctx, val.Message.Data); err != nil {
			// 	log.Fatal(err)
			// }
			ref := Conn.DBConn.NewRef("/")
			if err := ref.Set(ctx, val.Message.Data); err != nil {
				log.Fatalln("Error reading from database:", err)
			}

			go func() {
				time.Sleep(time.Minute * 5)
				postTweet(val.Message.Data.Text)
			}()
		}
	}
	return nil
}

func postTweet(text string) {
	tweet, _, err := client.Statuses.Update(text, nil)
	if err != nil {
		log.Print(err)
	}
	log.Print(tweet)
}

func init() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	config := oauth1.NewConfig(CONSUMER_KEY, CONSUMER_KEY_SECRET)
	token := oauth1.NewToken(ACCESS_TOKEN, ACCESS_SECRET)
	httpClient := config.Client(oauth1.NoContext, token)
	// Twitter client
	client = twitter.NewClient(httpClient)
	// for {
	// 	doEvery(2*time.Second, forFun)
	// }
	log.Print(dir)
	ctx := context.Background()
	opt := option.WithCredentialsFile(dir + "/twttr-bot-3dd9a-firebase-adminsdk-a9cni-2c1179fc4a.json")
	configDB := &firebase.Config{
		ProjectID:   "twttr-bot-3dd9a",
		DatabaseURL: "https://twttr-bot-3dd9a.firebaseio.com",
	}
	app, err := firebase.NewApp(context.Background(), configDB, opt)
	if err != nil {
		log.Print(fmt.Errorf("error initializing app: %v", err))
	}
	clientDB, err := app.Database(ctx)
	if err != nil {
		log.Fatalln("Error initializing database client:", err)
	}
	Conn = Connection{DBConn: clientDB}

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
