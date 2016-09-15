package main

import (
	"encoding/json"
	"errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	REQUEST_ERROR_TIMEOUT_SECONDS = 10
)

// NOTE: On a fresh Ubuntu (e.g. through docker) ca-certificate is not installed
// and making requests to https urls (like testributor's) is not possible, we
// see this error: https://github.com/zenazn/goji/issues/126
// git though seems to install ca-certificates as a dependency so it is not
// an issue since we always require git to be installed.
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

func SetupClientData() error {
	// Default url if environment var is not set
	if testributorUrl == "" {
		testributorUrl = "https://www.testributor.com/"
	}
	apiUrl = testributorUrl + "api/v1/"

	if appID == "" {
		return errors.New("APP_ID environment variable is not set.")
	}

	if appSecret == "" {
		return errors.New("APP_SECRET environment variable is not set.")
	}

	conf := &clientcredentials.Config{
		ClientID:     appID,
		ClientSecret: appSecret,
		//Scopes:       []string{"SCOPE1", "SCOPE2"},
		TokenURL: testributorUrl + "oauth/token",
	}
	tokenSource = conf.TokenSource(context.Background())

	return nil
}

type APIClient struct {
	http.Client
	logger Logger
}

// NewClient should be used to create an APIClient instance. A logger is required
// in order for the client to print the messages with the correct prefix.
func NewClient(logger Logger) *APIClient {
	return &APIClient{
		*oauth2.NewClient(context.Background(), tokenSource),
		logger,
	}
}

// HandleRequest takes an *http.Request, makes the request and returns the
// result as an empty interface.
// https://blog.golang.org/json-and-go
func (c *APIClient) PerformRequest(method string, path string, body string) (interface{}, error) {
	var request *http.Request
	var err error

	if body != "" {
		request, err = http.NewRequest(method, apiUrl+path, strings.NewReader(body))
		if err != nil {
			return nil, err
		}
	} else {
		request, err = http.NewRequest(method, apiUrl+path, nil)
		if err != nil {
			return nil, err
		}
	}

	if method == "POST" || method == "PATCH" {
		// multipart/form-data vs application/x-www-form-urlencoded
		// http://stackoverflow.com/a/4073451/974285
		request.Header.Add("Content-Type", "multipart/form-data")
	}
	request.Header.Add("WORKER_UUID", WorkerUUID)

	requestStart := time.Now()
	resp, err := c.Do(request)
	requestDuration := time.Since(requestStart).String()
	if err != nil {
		c.logger.Log("Error occured: " + err.Error())
		c.logger.Log("Error occured after " + requestDuration)
		c.logger.Log("Retrying in " + strconv.Itoa(REQUEST_ERROR_TIMEOUT_SECONDS) + " seconds")
		time.Sleep(REQUEST_ERROR_TIMEOUT_SECONDS * time.Second)
		return c.PerformRequest(method, path, body)
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
	if string(contents) != "" {
		err = json.Unmarshal(contents, &result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (c *APIClient) ProjectSetupData() (interface{}, error) {
	return c.PerformRequest("GET", "projects/setup_data", "")
}

func (c *APIClient) FetchJobs() (interface{}, error) {
	return c.PerformRequest("PATCH", "test_jobs/bind_next_batch", "")
}

func (c *APIClient) Beacon() (interface{}, error) {
	return c.PerformRequest("POST", "projects/beacon", "")
}

// http://codefol.io/posts/How-Does-Rack-Parse-Query-Params-With-parse-nested-query
func (c *APIClient) UpdateTestJobs(jobs []Job) (interface{}, error) {
	form := url.Values{}
	for _, job := range jobs {
		jobData, err := json.Marshal(job)
		if err != nil {
			return new(interface{}), err
		}
		form.Add("jobs["+job.GetId()+"]", string(jobData))
	}

	return c.PerformRequest("PATCH", "test_jobs/batch_update", form.Encode())
}
