package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type SalesforceResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Provide three arguments: Salesforce URL, client ID and client secret")
	}

	salesforceUrl := os.Args[1]
	clientId := os.Args[2]
	clientSecret := os.Args[3]

	log.Default().Println(clientSecret)

	r := gin.Default()

	r.GET("/auth/start", func(c *gin.Context) {
		startAuthFlow(c, salesforceUrl, clientId, clientSecret)
	})

	r.GET("/auth/complete", func(c *gin.Context) {
		requestAccessToken(c, salesforceUrl, clientId, clientSecret)
	})

	r.Run(":8888")
}

func startAuthFlow(c *gin.Context, salesforceUrl string, clientId string, clientSecret string) {
	url := salesforceUrl + "/services/oauth2/authorize"
	redirectUri := "http://localhost:8888/auth/complete"
	responseType := "code"
	prompt := "login consent"
	display := "popup"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating HTTP request: ", err)
	}

	q := req.URL.Query()
	q.Add("client_id", clientId)
	q.Add("redirect_uri", redirectUri)
	q.Add("response_type", responseType)
	q.Add("prompt", prompt)
	q.Add("display", display)
	req.URL.RawQuery = q.Encode()

	c.Redirect(302, req.URL.String())
}

func requestAccessToken(c *gin.Context, salesforceUrl string, clientId string, clientSecret string) {
	// Get authorization code from query parameters
	authCode := c.Request.URL.Query().Get("code")

	url := salesforceUrl + "/services/oauth2/token"
	grantType := "authorization_code"
	redirectUri := "http://localhost:8888/auth/complete"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Fatal("Error creating HTTP request: ", err)
	}

	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(clientId + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+encodedCredentials)

	q := req.URL.Query()
	q.Add("code", authCode)
	q.Add("grant_type", grantType)
	q.Add("redirect_uri", redirectUri)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error executing HTTP request: ", err)
	}
	defer resp.Body.Close()

	responseBody := SalesforceResponse{}
	json.NewDecoder(resp.Body).Decode(&responseBody)

	c.JSON(200, gin.H{
		"access_token":  responseBody.AccessToken,
		"refresh_token": responseBody.RefreshToken,
	})
}
