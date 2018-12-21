package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var config struct {
	TeamName      string
	Email         string
	Password      string
	Token         string
	Quiet         bool
	NotifyChannel string
	Loop          bool
}

var (
	cookieJar, _      = cookiejar.New(nil)
	client            = &http.Client{Jar: cookieJar}
	everyTwoSeconds   <-chan time.Time
	everyThreeSeconds <-chan time.Time
)

var (
	crumbRegex    = regexp.MustCompile(`name="crumb" value="([^\"]+)"`)
	apiTokenRegex = regexp.MustCompile(`api_token\s*:\s*"([^\"]+)"`)
)

var (
	domain  string
	baseURL string
)

func notificationMessageJSON(emojiNames []string) string {
	buf := bytes.NewBuffer(nil)
	for _, emojiName := range emojiNames {
		fmt.Fprintf(buf, ":%s: ", emojiName)
	}
	return fmt.Sprintf(`{ "channel": %q, "text": "> 🤖 *emoji-bot*\n> %s" }`, config.NotifyChannel, buf.String())
}

func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime)
	flag.StringVar(&config.TeamName, "team", "", "Slack team (only needed when using email/password auth)")
	flag.StringVar(&config.NotifyChannel, "notify-channel", "", "Notify this channel on successful uploads")
	flag.StringVar(&config.Token, "token", os.Getenv("SLACK_API_TOKEN"), "Slack API token")
	flag.StringVar(&config.Email, "email", "", "user email")
	flag.StringVar(&config.Password, "password", "", "user password")
	flag.BoolVar(&config.Quiet, "quiet", false, "suppress log output")
	flag.BoolVar(&config.Loop, "loop", false, "repeatedly sync")
	flag.Parse()

	if config.Quiet {
		log.SetOutput(ioutil.Discard)
	}

	if config.Token != "" && config.Email != "" {
		config.TeamName = "api"
	}

	if config.TeamName == "" {
		log.Fatal("required parameter: -team")
	}

	domain = fmt.Sprintf("%s.slack.com", config.TeamName)
	baseURL = fmt.Sprintf("https://%s", domain)

	if config.Email != "" {
		apiToken, err := obtainToken()
		if err != nil {
			log.Fatalf("email/password auth failed: %v", err)
		}
		config.Token = apiToken
		log.Printf("obtained token: %q", config.Token)
	}
	if config.Token == "" {
		log.Fatal("required parameters: -token or -email/-password")
	}

	everyTwoSeconds = time.Tick(2 * time.Second)
	everyThreeSeconds = time.Tick(3 * time.Second)
}

func obtainToken() (string, error) {
	// Get CSRF-protection token ("crumb")
	req, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return "", err
	}

	<-everyTwoSeconds
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	crumbs := crumbRegex.FindSubmatch(respBytes)
	if len(crumbs) < 2 {
		return "", fmt.Errorf("no crumbs")
	}
	crumb := string(crumbs[1])

	// Log in and scrape an API token
	form := url.Values{
		"email":        {config.Email},
		"password":     {config.Password},
		"crumb":        {crumb},
		"signin":       {"1"},
		"redir":        {""},
		"has_remember": {"1"},
	}
	req, err = http.NewRequest(http.MethodPost, baseURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	<-everyTwoSeconds
	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	respBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	apiTokens := apiTokenRegex.FindSubmatch(respBytes)
	if len(apiTokens) < 2 {
		return "", fmt.Errorf("could not scrape API token")
	}
	apiToken := string(apiTokens[1])

	return apiToken, nil
}

func listEmoji() (map[string]string, error) {
	apiURL := fmt.Sprintf("%s/api/emoji.list", baseURL)
	req, _ := http.NewRequest(http.MethodGet, apiURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Token))

	<-everyThreeSeconds
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResponse struct {
		Ok    bool              `json:"ok"`
		Error string            `json:"error"`
		Emoji map[string]string `json:"emoji"`
	}
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		return nil, err
	}

	if !apiResponse.Ok {
		return nil, fmt.Errorf("%s", apiResponse.Error)
	}

	return apiResponse.Emoji, nil
}

func notifyEmojiUploaded(messageJSON string) error {
	if config.NotifyChannel == "" {
		return nil
	}

	apiURL := fmt.Sprintf("%s/api/chat.postMessage", baseURL)
	req, _ := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(messageJSON))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Token))
	req.Header.Set("Content-Type", "application/json")

	<-everyTwoSeconds
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var apiResponse struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	}

	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		return err
	}
	if !apiResponse.Ok {
		return fmt.Errorf("%s", apiResponse.Error)
	}
	return nil
}

func uploadEmoji(fileName, emojiName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	apiURL := fmt.Sprintf("%s/api/emoji.add", baseURL)

	body := bytes.NewBuffer(nil)
	bodyWriter := multipart.NewWriter(body)

	bodyWriter.WriteField("mode", "data")
	bodyWriter.WriteField("name", emojiName)
	image, _ := bodyWriter.CreateFormFile("image", filepath.Base(fileName))
	io.Copy(image, f)
	bodyWriter.WriteField("token", config.Token)
	bodyWriter.Close()

	req, err := http.NewRequest(http.MethodPost, apiURL, body)
	if err != nil {
		return err
	}

	req.Header.Set("Origin", baseURL)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	<-everyTwoSeconds
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		return fmt.Errorf("HTTP %d: %q", resp.StatusCode, bodyString)
	}

	var apiResponse struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	}

	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		return err
	}
	if !apiResponse.Ok {
		return fmt.Errorf("%s", apiResponse.Error)
	}
	return nil
}

func main() {
	files := flag.Args()
	for {
		uploadFilesAndPrintSummary(files)
		if !config.Loop {
			return
		}
	}
}

func uploadFilesAndPrintSummary(files []string) {
	var (
		summary   = map[string][]string{}
		summaryMu sync.Mutex
	)
	var wg sync.WaitGroup
	log.Println("fetching emoji list")
	currentEmoji, err := listEmoji()
	if err != nil {
		log.Println(err)
		return
	}

	const skipKey = "skip"
	const successKeyRestored = "successRestored"
	const successKey = "successAdded"

	for _, filePath := range files {
		ext := filepath.Ext(filePath)
		emojiName := strings.TrimSuffix(filepath.Base(filePath), ext)
		if _, ok := currentEmoji[emojiName]; ok {
			log.Printf("%s: already exists, skipping", emojiName)
			summary[skipKey] = append(summary[skipKey], emojiName)
			continue
		}
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			log.Printf("%s: uploading %q", emojiName, filePath)
			err := uploadEmoji(filePath, emojiName)
			if err != nil {
				error := err.Error()
				summaryMu.Lock()
				summary[error] = append(summary[error], emojiName)
				summaryMu.Unlock()
				return
			}

			summaryMu.Lock()
			summary[successKey] = append(summary[successKey], emojiName)
			summaryMu.Unlock()
		}(filePath)
	}
	wg.Wait()

	if len(summary[successKey]) > 0 {
		err = notifyEmojiUploaded(notificationMessageJSON(summary[successKey]))
		if err != nil {
			log.Printf("notification failed: %v", err)
		}
	}

	var output struct {
		Emoji  map[string][]string `json:",omitempty"`
		Counts map[string]int
	}
	output.Emoji = summary
	output.Counts = map[string]int{}
	for k, v := range summary {
		sort.Strings(v)
		output.Counts[k] = len(v)
	}
	if len(output.Emoji) == 1 {
		if _, ok := output.Emoji[skipKey]; ok {
			output.Emoji = nil
		}
	}
	json.NewEncoder(os.Stdout).Encode(output)
}
