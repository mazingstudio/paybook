package paybook

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	apiPrefix    = "https://sync.paybook.com/v1"
	staticPrefix = "https://s.paybook.com"
)

type Time time.Time

type StatusCodes []StatusCode

func (sc StatusCodes) Last() int {
	v := []StatusCode(sc)
	if len(v) < 1 {
		return 1
	}
	return v[len(v)-1].Code
}

func (t *Time) UnmarshalJSON(data []byte) error {
	ut, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*t = Time(time.Unix(ut, 0))
	return nil
}

type Account struct {
}

type Transaction struct {
}

type SiteOrganization struct {
	IDSiteOrganization     string `json:"id_site_organization"`
	IDSiteOrganizationType string `json:"id_site_organization_type"`
	IDCountry              string `json:"id_country"`
	Name                   string `json:"name"`
	Avatar                 string `json:"avatar"`
	SmallCover             string `json:"small_cover"`
	Cover                  string `json:"cover"`
}

type StatusCode struct {
	Code int `json:"code"`
}

type CredentialRequest struct {
	IDSite      string            `json:"id_site"`
	Credentials map[string]string `json:"credentials"`
	Token       string            `json:"token"`
}

type AccountCredential struct {
	IDCredential string `json:"id_credential"`
	Username     string `json:"username"`
	WS           string `json:"ws"`
	Status       string `json:"status"`
	TFA          string `json:"twofa"`
}

type Envelope struct {
	RID     string      `json:"rid"`
	Code    int         `json:"code"`
	Status  bool        `json:"status"`
	Errors  interface{} `json:"errors"`
	Message *string     `json:"message"`
}

type User struct {
	Name       string `json:"name"`
	ID         string `json:"id_user,omitempty"`
	CreatedAt  *Time  `json:"dt_create"`
	ModifiedAt *Time  `json:"dt_modify"`
}

type Session struct {
	Token string `json:"token"`
	Key   string `json:"key"`
	IV    string `json:"iv"`
}

type Client struct {
	APIKey     string
	httpClient *http.Client
}

type Credential struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Label      string      `json:"label"`
	Required   bool        `json:"required"`
	Username   bool        `json:"username"`
	Validation interface{} `json:"validation,omitempty"`
}

type Catalogue struct {
	IDSite                 string       `json:"id_site"`
	IDSiteOrganization     string       `json:"id_site_organization"`
	IDSiteOrganizationType string       `json:"id_site_organization_type"`
	Name                   string       `json:"name"`
	Credentials            []Credential `json:"credentials"`
}

func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("Missing API key")
	}
	return &Client{
		httpClient: &http.Client{}, // TODO: timeout, etc.
		APIKey:     apiKey,
	}, nil
}

func (c *Client) CreateUser(user *User) (*User, error) {
	res := struct {
		Envelope
		Response *User `json:"response,omitempty"`
	}{}

	err := c.post("/users", user, &res)
	if err != nil {
		return nil, err
	}

	if !res.Status {
		return nil, errors.New(*res.Message)
	}

	return res.Response, nil
}

func (c *Client) CreateSession(user *User) (*Session, error) {
	res := struct {
		Envelope
		Response *Session `json:"response,omitempty"`
	}{}

	err := c.post("/sessions", user, &res)
	if err != nil {
		return nil, err
	}

	if !res.Status {
		return nil, errors.New(*res.Message)
	}

	return res.Response, nil
}

func (c *Client) Transactions(params url.Values) ([]Transaction, error) {
	res := struct {
		Envelope
		Response []Transaction `json:"response,omitempty"`
	}{}

	err := c.get("/transactions", params, &res)
	if err != nil {
		return nil, err
	}

	if !res.Status {
		return nil, errors.New(*res.Message)
	}

	return res.Response, nil
}

func (c *Client) Accounts(params url.Values) ([]Account, error) {
	res := struct {
		Envelope
		Response []Account `json:"response,omitempty"`
	}{}

	err := c.get("/accounts", params, &res)
	if err != nil {
		return nil, err
	}

	if !res.Status {
		return nil, errors.New(*res.Message)
	}

	return res.Response, nil
}

func (c *Client) Status(statusURL string, params url.Values) (StatusCodes, error) {
	res := struct {
		Envelope
		Response StatusCodes `json:"response,omitempty"`
	}{}

	err := c.get(statusURL, params, &res)
	if err != nil {
		return nil, err
	}

	if !res.Status {
		return nil, errors.New(*res.Message)
	}

	return res.Response, nil
}

func (c *Client) CreateCredential(user *CredentialRequest) (*AccountCredential, error) {
	res := struct {
		Envelope
		Response *AccountCredential `json:"response,omitempty"`
	}{}

	err := c.post("/credentials", user, &res)
	if err != nil {
		return nil, err
	}

	if !res.Status {
		return nil, errors.New(*res.Message)
	}

	return res.Response, nil
}

func (c *Client) SiteOrganizations() ([]SiteOrganization, error) {
	res := struct {
		Envelope
		Response []SiteOrganization `json:"response,omitempty"`
	}{}

	err := c.get("/catalogues/site_organizations", nil, &res)
	if err != nil {
		return nil, err
	}

	return res.Response, nil
}

func (c *Client) Catalogues(params url.Values) ([]Catalogue, error) {
	res := struct {
		Envelope
		Response []Catalogue `json:"response,omitempty"`
	}{}

	err := c.get("/catalogues/sites", params, &res)
	if err != nil {
		return nil, err
	}

	return res.Response, nil
}

func (c *Client) get(endpoint string, params url.Values, dest interface{}) error {
	res, err := c.httpClient.Get(c.signEndpoint(endpoint, params))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	log.Printf("GOT (%v): %v", endpoint, string(out))

	if err := json.Unmarshal(out, dest); err != nil {
		return err
	}

	return nil
}

func (c *Client) post(endpoint string, data interface{}, dest interface{}) error {
	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}
	res, err := c.httpClient.Post(c.signEndpoint(endpoint, nil), "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(out, dest); err != nil {
		return err
	}

	return nil
}

func (c *Client) signEndpoint(endpoint string, params url.Values) string {
	if params == nil {
		params = url.Values{}
	}

	if params.Get("api_key") == "" {
		params.Set("api_key", c.APIKey)
	}

	fqu := endpoint
	if strings.HasPrefix(fqu, "/") {
		fqu = apiPrefix + endpoint
	}

	uri, err := url.Parse(fqu)
	if err != nil {
		panic(err.Error())
	}

	uri.RawQuery = params.Encode()
	return uri.String()
}

func AssetURL(name string) string {
	return staticPrefix + name
}
