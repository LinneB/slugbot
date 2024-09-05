package helix

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Helix struct {
	LiveChannels map[string]bool
	channels     []string
	clientID     string
	clientSecret string
	debug        bool
	httpClient   *http.Client
	token        string
}

type AuthError struct{}

func (ae *AuthError) Error() string {
	return "Helix authentication error"
}

func New(clientid string, secret string, channels []string, debug bool) (*Helix, error) {
	c := Helix{
		LiveChannels: make(map[string]bool),
		channels:     channels,
		clientID:     clientid,
		clientSecret: secret,
		debug:        debug,
		httpClient:   &http.Client{},
	}
	token, err := c.generateToken()
	if err != nil {
		return nil, fmt.Errorf("Could not generate token: %w", err)
	}
	c.token = token
	go func() {
		for {
			c.DebugLog("Updating live channels")
			err := c.updateLiveChannels()
			if err != nil {
				var authError *AuthError
				if errors.As(err, &authError) {
					log.Println("Token is invalid, generating a new one...")
					token, err := c.generateToken()
					if err != nil {
						log.Fatalf("Invalid clientid/secret")
					}
					log.Println("Succesfully generated a new token")
					c.token = token
					continue
				} else {
					log.Fatalf("%s", err)
				}
			}
			time.Sleep(60 * time.Second)
		}
	}()
	return &c, nil
}

func (c *Helix) DebugLog(format string, a ...any) {
	if c.debug {
		log.Printf(format, a...)
	}
}

func (c *Helix) generateToken() (string, error) {
	values := url.Values{
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"grant_type":    {"client_credentials"},
	}

	res, err := http.PostForm("https://id.twitch.tv/oauth2/token", values)
	if err != nil {
		return "", err
	}
	if res.StatusCode == 400 {
		return "", fmt.Errorf("Helix returned 400: ClientID/secret invalid")
	}

	var jsonBody struct {
		AccessToken string `json:"access_token"`
	}
	if err = json.NewDecoder(res.Body).Decode(&jsonBody); err != nil {
		return "", err
	}
	return jsonBody.AccessToken, nil
}

func (c *Helix) updateLiveChannels() error {
	output := make(map[string]bool)
	for _, channels := range chunkArray(c.channels, 100) {
		url := fmt.Sprintf("https://api.twitch.tv/helix/streams?user_login=%s", strings.Join(channels, "&user_login="))
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		request.Header.Add("Client-ID", c.clientID)
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
		res, err := c.httpClient.Do(request)
		if err != nil {
			return err
		}

		if res.StatusCode == 401 {
			return &AuthError{}
		}
		if res.StatusCode != 200 {
			return fmt.Errorf("Helix returned unhandled status code")
		}

		var jsonBody struct {
			Data []struct {
				UserLogin string `json:"user_login"`
			} `json:"data"`
		}
		err = json.NewDecoder(res.Body).Decode(&jsonBody)
		if err != nil {
			return err
		}
		for _, stream := range jsonBody.Data {
			c.DebugLog("%s is live", stream.UserLogin)
			output[stream.UserLogin] = true
		}
	}
	c.LiveChannels = output
	return nil
}

// Splits array into arrays of size chunkSize
func chunkArray[V any](array []V, chunkSize int) [][]V {
	var output [][]V
	for i := 0; i < len(array); i += chunkSize {
		j := min(i+chunkSize, len(array))
		chunk := array[i:j]
		output = append(output, chunk)
	}
	return output
}
