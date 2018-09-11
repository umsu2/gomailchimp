package mailchimp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"net/http"
	"net/url"
	"strings"
)

// The MailChimp API url structure.
const ApiURL string = "https://%s.api.mailchimp.com/3.0/"
const AccountRoute string = ""

var ErrAPIKeyFormat = errors.New("Invalid API key format")

type client struct {
	httpClient *http.Client
	credential credential
	baseUrl    *url.URL
}

//todo fill this up.
type AccountInfo struct {
	AccountName string `json:"account_name"`
	AccountId   string `json:"account_id"`
	LoginId     string `json:"login_id"`
}

type accountServiceOp struct {
	client *client
}

func (s *accountServiceOp) GetAccountInfo(ctx context.Context) (AccountInfo, error) {
	accountInfo := AccountInfo{}
	err := s.client.Get(ctx, AccountRoute, &accountInfo, nil)
	return accountInfo, err
}

type AccountService interface {
	GetAccountInfo(ctx context.Context) (AccountInfo, error)
}
type ChimpApi interface {
	GetAccountService(ctx context.Context) AccountService
	GetClient(ctx context.Context) Client
	//GetClient() Client
	//GetListService() ListService
	//GetMemberService() MemberService
}

type credential struct {
	ApiKey string
	Region string
}
type chimpApiOp struct {
	client *client
}

func (c *chimpApiOp) GetAccountService(ctx context.Context) AccountService {
	return &accountServiceOp{c.client}
}
func (c *chimpApiOp) GetClient(ctx context.Context) Client {
	return c.client
}

type Client interface {
	Do(ctx context.Context, req *http.Request, v interface{}) error
	NewRequest(ctx context.Context, method, urlStr string, body, options interface{}) (*http.Request, error)
	CreateAndDo(ctx context.Context, method, path string, data, options, resource interface{}) error
	Get(ctx context.Context, path string, resource, options interface{}) error
	Post(ctx context.Context, path string, data, resource interface{}) error
	Put(ctx context.Context, path string, data, resource interface{}) error
	Delete(ctx context.Context, path string) error
}

// Creates an API request. A relative URL can be provided in urlStr, which will
// be resolved to the BaseURL of the Client. Relative URLS should always be
// specified without a preceding slash.
func (c *client) NewRequest(ctx context.Context, method, urlStr string, body, options interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	// Make the full url based on the relative path
	u := c.baseUrl.ResolveReference(rel)

	// Add query params, if provided
	if options != nil {
		optionsQuery, err := query.Values(options)
		if err != nil {
			return nil, err
		}

		for k, values := range u.Query() {
			for _, v := range values {
				optionsQuery.Add(k, v)
			}
		}
		u.RawQuery = optionsQuery.Encode()
	}

	// A bit of JSON ceremony
	var byt []byte = nil

	if body != nil {
		byt, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), bytes.NewBuffer(byt))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth("", c.credential.ApiKey)

	return req, nil
}

// CreateAndDo performs a web request to MailChimp with the given method (GET,
// POST, PUT, DELETE) and relative path
func (c *client) CreateAndDo(ctx context.Context, method, path string, data, options, resource interface{}) error {
	req, err := c.NewRequest(ctx, method, path, data, options)
	if err != nil {
		return err
	}
	err = c.Do(ctx, req, resource)
	if err != nil {
		return err
	}
	return nil
}

// Do sends an API request and populates the given interface with the parsed
// response.
func (c *client) Do(ctx context.Context, req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = CheckResponseError(resp)
	if err != nil {
		return err
	}

	if v != nil {
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&v)
		if err != nil {
			return err
		}
	}

	return nil
}
func CheckResponseError(r *http.Response) error {
	// need to parse the request to see if error has occured.
	return nil
}

// Get performs a GET request for the given path and saves the result in the
// given resource.
func (c *client) Get(ctx context.Context, path string, resource, options interface{}) error {
	return c.CreateAndDo(ctx, "GET", path, nil, options, resource)
}

// Post performs a POST request for the given path and saves the result in the
// given resource.
func (c *client) Post(ctx context.Context, path string, data, resource interface{}) error {
	return c.CreateAndDo(ctx, "POST", path, data, nil, resource)
}

// Put performs a PUT request for the given path and saves the result in the
// given resource.
func (c *client) Put(ctx context.Context, path string, data, resource interface{}) error {
	return c.CreateAndDo(ctx, "PUT", path, data, nil, resource)
}

// Delete performs a DELETE request for the given path
func (c *client) Delete(ctx context.Context, path string) error {
	return c.CreateAndDo(ctx, "DELETE", path, nil, nil, nil)
}

func newCredential(ApiKey string) (credential, error) {

	split := strings.Split(ApiKey, "-")
	if len(split) != 2 {
		return credential{}, ErrAPIKeyFormat
	}
	return credential{
		ApiKey: ApiKey,
		Region: split[1],
	}, nil

}

func NewChimpApi(apiKey string) (ChimpApi, error) {
	credential, err := newCredential(apiKey)
	if err != nil {
		return nil, err
	}
	baseUrl, err := url.Parse(fmt.Sprintf(ApiURL, credential.Region))
	if err != nil {
		return nil, err
	}
	client := client{
		httpClient: http.DefaultClient,
		credential: credential,
		baseUrl:    baseUrl,
	}
	return &chimpApiOp{
		client: &client,
	}, nil
}
