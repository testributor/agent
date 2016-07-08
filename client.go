package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"net/http"
	"os"
)

var testributorUrl = os.Getenv("TESTRIBUTOR_URL")
var apiUrl string
var appID = os.Getenv("APP_ID")
var appSecret = os.Getenv("APP_SECRET")

// We build only one tokenSource and use it to create new APIClients
// (through NewClient() function) to avoid making multiple requests for
// token generation.
// oauth2.TokenSource is an interface which if implemented in our case by a
// *oauth2.reuseTokenSource (Use fmt.Printf("%T\n", tokenSource) to see that).
// Since this is a pointer, only the first call to
// NewClient() will perform a token request. After that the token will be reused
// since it points to an already initialized TokenSource value (with a valid token).
var tokenSource oauth2.TokenSource

func init() {
	// Default url if environment var is not set
	if testributorUrl == "" {
		testributorUrl = "https://www.testributor.com/"
	}
	apiUrl = testributorUrl + "api/v1/"

	if appID == "" {
		fmt.Println("APP_ID environment variable is not set. Exiting.")
		os.Exit(1)
	}
	if appSecret == "" {
		fmt.Println("APP_SECRET environment variable is not set. Exiting.")
		os.Exit(1)
	}

	conf := &clientcredentials.Config{
		ClientID:     appID,
		ClientSecret: appSecret,
		//Scopes:       []string{"SCOPE1", "SCOPE2"},
		TokenURL: testributorUrl + "oauth/token",
	}
	tokenSource = conf.TokenSource(context.Background())
}

func NewClient() *APIClient {
	return &APIClient{
		*oauth2.NewClient(context.Background(), tokenSource),
	}
}

type APIClient struct {
	http.Client
}

// HandleRequest takes an *http.Request, makes the request and returns the
// result as an empty interface.
// https://blog.golang.org/json-and-go
func (c *APIClient) PerformRequest(method string, path string) (interface{}, error) {
	request, err := http.NewRequest(method, apiUrl+path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		return nil, errors.New("Authentication error")
	}

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal(contents, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *APIClient) ProjectSetupData() (interface{}, error) {
	return c.PerformRequest("GET", "projects/setup_data")
}

func (c *APIClient) FetchJobs() (interface{}, error) {
	return c.PerformRequest("PATCH", "test_jobs/bind_next_batch")
}
