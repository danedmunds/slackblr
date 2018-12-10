package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bakerolls/gotumblr"
)

var allowedUserIds map[string]string
var tumblr *gotumblr.Client
var blog string
var slackSigningSecret string
var slackToken string
var channel string
var command string
var username string

func main() {
	// slack
	err := json.Unmarshal([]byte(envVar("SLACK_USERS")), &allowedUserIds)
	if err != nil {
		panic("SLACK_USERS must be valid json")
	}
	slackSigningSecret = envVar("SLACK_SIGNING_SECRET")
	slackToken = envVar("SLACK_TOKEN")
	channel = envVar("SLACK_CHANNEL")
	command = envVar("SLACK_COMMAND")
	username = envVar("SLACK_USERNAME")

	// tumblr
	blog = envVar("TUMBLR_BLOG")
	key := envVar("TUMBLR_KEY")
	secret := envVar("TUMBLR_SECRET")
	token := envVar("TUMBLR_TOKEN")
	tokenSecret := envVar("TUMBLR_TOKEN_SECRET")
	tumblr = gotumblr.New(key, secret, token, tokenSecret)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	http.HandleFunc("/", slackHook)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func envVar(name string) (value string) {
	value = os.Getenv(name)
	if value == "" {
		panic(fmt.Sprintf("Missing required environment variable %s", name))
	}
	return
}

func slackHook(w http.ResponseWriter, r *http.Request) {
	message := "Trying to post to tumblr..."
	defer func() {
		r := recover()
		if r != nil {
			message = r.(string)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"response_type": "ephemeral",
			"text":          message,
		})
	}()

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(fmt.Errorf("Failed to read request body: %s", err.Error()))
	}
	body := string(bodyBytes)
	verifyFromSlack(r, body)

	form, err := url.ParseQuery(body)
	if err != nil {
		panic("Invalid request")
	}

	user := form.Get("user_id")
	requestedCommand := form.Get("command")
	text := form.Get("text")
	resURL := form.Get("response_url")

	name := allowedUserIds[user]
	if name == "" {
		panic("User not allowed")
	}

	if requestedCommand != command {
		panic("Invalid command")
	}

	parsed, err := url.Parse(strings.Trim(text, " \t\n\r"))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		panic("You must send a full URL")
	}

	go sendToTumblr(parsed, resURL, name)
}

func verifyFromSlack(r *http.Request, body string) error {
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")
	signature := r.Header.Get("X-Slack-Signature")
	message := fmt.Sprintf("v0:%s:%s", timestamp, body)

	now := time.Now().Unix()
	unixTime, err := strconv.Atoi(timestamp)
	if err != nil || int64(unixTime+5*60) < now {
		return fmt.Errorf("Request is too old, potential replay attack")
	}

	mac := hmac.New(sha256.New, []byte(slackSigningSecret))
	mac.Write([]byte(message))
	computedSignature := fmt.Sprintf("v0=%s", hex.EncodeToString(mac.Sum(nil)))

	if computedSignature != signature {
		return fmt.Errorf("Message signature is invalid")
	}

	return nil
}

func sendToTumblr(parsedURL *url.URL, resURL, tag string) {
	defer func() {
		r := recover()
		if r != nil {
			err := r.(error)
			fmt.Printf("Failed to post to tumblr: %s\n", err.Error())
		}
	}()

	res, err := http.Head(parsedURL.String())
	handle(err)

	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		panic(fmt.Errorf("Wrong content type, got %s", contentType))
	}

	err = tumblr.CreatePhoto(blog, url.Values{
		"source":  {parsedURL.String()},
		"tags":    {tag},
		"caption": {fmt.Sprintf("Source: %s", parsedURL.String())},
	})
	handle(err)

	err = sendToChannel(channel, parsedURL.String())
	handle(err)

	err = respond(resURL, "Posted!")
	if err != nil {
		fmt.Printf("Failed to respond to Slack: %s", err.Error())
	}
}

func sendToChannel(channel, text string) error {
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", strings.NewReader(url.Values{
		"channel":  {channel},
		"text":     {text},
		"username": {username},
	}.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", slackToken))

	_, err = http.DefaultClient.Do(req)
	return err
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func respond(responseURL, message string) error {
	body, _ := json.Marshal(map[string]interface{}{
		"response_type": "ephemeral",
		"text":          message,
	})
	res, err := http.Post(responseURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return err
	}

	if res.StatusCode > 299 && res.StatusCode < 200 {
		return fmt.Errorf("Unexpected response code: %d", res.StatusCode)
	}

	return nil
}
