package paybook

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
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

type asset string

func (a *asset) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "" {
		return nil
	}
	var assetURL string
	if err := json.Unmarshal(data, &assetURL); err != nil {
		return err
	}
	*a = asset(staticPrefix + assetURL)
	return nil
}

type Account struct {
	AccountType string  `json:"account_type"`
	Balance     float64 `json:"balance"`
	Currency    string  `json:"currency"`
	RefreshedAt *Time   `json:"dt_refresh"`
	//Extra              interface{}      `json:"extra"`
	IDAccount     string `json:"id_account"`
	IDAccountType string `json:"id_account_type"`
	IDCredential  string `json:"id_credential"`
	//IDExternal         interface{}      `json:"id_external"`
	IDSite             string           `json:"id_site"`
	IDSiteOrganization string           `json:"id_site_organization"`
	IDUser             string           `json:"id_user"`
	IsDisable          int              `json:"is_disable"`
	Name               string           `json:"name"`
	Number             string           `json:"number"`
	Site               SiteOrganization `json:"site"`
}

type Attachment struct {
	IDAttachment     string `json:"id_attachment"`
	IDAttachmentType string `json:"id_attachment_type"`
	IsValid          int    `json:"is_valid"`
	File             string `json:"file"`
	MIME             string `json:"mime"`
	URL              asset  `json:"url"`
}

type Transaction struct {
	IDTransaction          string       `json:"id_transaction"`
	IDUser                 string       `json:"id_user"`
	IDSite                 string       `json:"id_site"`
	IDSiteOrganization     string       `json:"id_site_organization"`
	IDSiteOrganizationType string       `json:"id_site_organization_type"`
	IDAccount              string       `json:"id_account"`
	IDAccountType          string       `json:"id_account_type"`
	IDCurrency             string       `json:"id_currency"`
	IsDisable              int          `json:"is_disable"`
	Amount                 float64      `json:"amount"`
	Currency               string       `json:"currency"`
	Reference              interface{}  `json:"reference"`
	Keywords               interface{}  `json:"keywords"`
	Extra                  interface{}  `json:"extra"`
	Attachments            []Attachment `json:"attachments"`
	CreatedAt              *Time        `json:"dt_transaction"`
	RefresedAt             *Time        `json:"dt_refresh"`
	DisabledAt             *Time        `json:"dt_disable"`
	Description            string       `json:"description"`
}

type SiteOrganization struct {
	IDSiteOrganization     string `json:"id_site_organization"`
	IDSiteOrganizationType string `json:"id_site_organization_type"`
	IDCountry              string `json:"id_country"`
	Name                   string `json:"name"`
	Avatar                 asset  `json:"avatar"`
	SmallCover             asset  `json:"small_cover"`
	Cover                  asset  `json:"cover"`
	Organization           string `json:"organization"`
	TimeZone               string `json:"time_zone"`
}

type StatusCode struct {
	Code int `json:"code"`
}

type CredentialRequest struct {
	IDSite      string            `json:"id_site"`
	IDUser      string            `json:"id_user"`
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

	err := c.post("/users", nil, user, &res)
	if err != nil {
		return nil, err
	}

	if !res.Status {
		return nil, errors.New(*res.Message)
	}

	return res.Response, nil
}

func (c *Client) Users(params url.Values) ([]User, error) {
	res := struct {
		Envelope
		Response []User `json:"response,omitempty"`
	}{}

	err := c.get("/users", params, &res)
	if err != nil {
		return nil, err
	}

	if !res.Status {
		return nil, errors.New(*res.Message)
	}

	if len(res.Response) < 1 {
		return nil, errors.New("empty result")
	}

	return res.Response, nil
}

func (c *Client) CreateSession(user *User) (*Session, error) {
	res := struct {
		Envelope
		Response *Session `json:"response,omitempty"`
	}{}

	err := c.post("/sessions", nil, user, &res)
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

func (c *Client) RemoveToken(token string) (bool, error) {
	res := struct {
		Envelope
		Response bool `json:"response,omitempty"`
	}{}
	err := c.delete("/sessions/"+token, nil, &res)
	if err != nil {
		return false, err
	}
	return res.Response, nil
}

func (c *Client) ValidToken(token string) (bool, error) {
	res := struct {
		Envelope
		Response bool `json:"response,omitempty"`
	}{}

	err := c.get("/sessions/"+token+"/verify", nil, &res)
	if err != nil {
		return false, err
	}

	if !res.Status {
		return false, errors.New(*res.Message)
	}

	return res.Response, nil
}

func (c *Client) CreateCredential(user *CredentialRequest) (*AccountCredential, error) {
	res := struct {
		Envelope
		Response *AccountCredential `json:"response,omitempty"`
	}{}

	err := c.post("/credentials", nil, user, &res)
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

	if err := json.Unmarshal(out, dest); err != nil {
		return err
	}

	return nil
}

func (c *Client) post(endpoint string, params url.Values, data interface{}, dest interface{}) error {
	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}
	res, err := c.httpClient.Post(c.signEndpoint(endpoint, params), "application/json", bytes.NewBuffer(buf))
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

func (c *Client) delete(endpoint string, params url.Values, dest interface{}) error {
	req, err := http.NewRequest("DELETE", c.signEndpoint(endpoint, params), nil)
	if err != nil {
		return err
	}

	res, err := c.httpClient.Do(req)
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
