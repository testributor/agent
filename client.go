package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	// TODO: Change this to some standard location (User's home?)
	TOKEN_FILE = ".testributor"
)

var apiUrl = os.Getenv("TESTRIBUTOR_API_URL")
var conf = &oauth2.Config{
	//Scopes:       []string{},
	Endpoint: oauth2.Endpoint{TokenURL: apiUrl + "oauth/token"},
}

func token() *oauth2.Token {
	var token oauth2.Token

	if _, err := os.Stat(TOKEN_FILE); err == nil {
		data, err := ioutil.ReadFile(TOKEN_FILE)
		check(err)
		json.Unmarshal(data, &token)
		check(err)

		return &token
	} else {
		fmt.Print("Email: ")
		username := ""
		fmt.Scanln(&username)

		fmt.Print("Password: ")
		password := ""
		fmt.Scanln(&password)

		defer func() {
			if rec := recover(); rec != nil {
				fmt.Println(rec)
				os.Exit(1)
			}
		}()
		token, err :=
			conf.PasswordCredentialsToken(context.Background(), username, password)
		check(err)

		data, err := json.Marshal(token)
		check(err)
		err = ioutil.WriteFile(TOKEN_FILE, data, 0644)
		check(err)

		return token
	}
}

func init() {
	// Default url if environment var is not set
	if apiUrl == "" {
		apiUrl = "https://testributor.herokuapp.com/"
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func NewClient() *APIClient {
	return (*APIClient)(conf.Client(context.Background(), token()))
}

type RequestHandler interface {
	HandleRequest(r *http.Request) string
}

type APIClient http.Client

// HandleRequest takes an *http.Request, makes the request and returns the
// result as a string. If we get a 401 response, we assume that the token
// might not be valid so we ask the user to generate a new one by providing
// her credentials again.
func (c *APIClient) HandleRequest(request *http.Request) string {
	// TODO: This makes 2 requests. One to exchange credentials for a token
	// and one for the actual request itself. Consider caching the token
	// (in a file?) and only requesting a new one when that one expires (if ever).
	resp, err := (*http.Client)(c).Do(request)
	check(err)

	if resp.StatusCode == 401 {
		fmt.Println("No valid token present. Let's generate a new one.")
		err = os.Remove(TOKEN_FILE)
		check(err)

		// Replace the client with a new one. This will call token() and will write
		// the TOKEN_FILE.
		*c = *NewClient()
		resp, err = (*http.Client)(c).Do(request)
		check(err)
	}

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return resp.Status + ": " + string(contents)
}

func checkCommitStatus(handler RequestHandler, commitHash string, projectName string) string {
	// TODO: Url encode projectName
	request, err := http.NewRequest("GET",
		apiUrl+"users_api/v1/commits/"+commitHash+"/status?project="+projectName, nil)

	if err != nil {
		fmt.Println(err)
	}

	return handler.HandleRequest(request)
}

func account(handler RequestHandler) string {
	request, err :=
		http.NewRequest("GET", apiUrl+"users_api/v1/users/current", nil)
	if err != nil {
		fmt.Println(err)
	}

	return handler.HandleRequest(request)
}
