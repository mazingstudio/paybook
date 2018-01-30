package paybook

import (
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var client *Client

func TestNewClient(t *testing.T) {
	var err error
	client, err = NewClient(os.Getenv("API_KEY"))
	assert.NoError(t, err)
}

func TestCreateEmptyUser(t *testing.T) {
	user, err := client.CreateUser(&User{})
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestCreateUser(t *testing.T) {
	user, err := client.CreateUser(&User{
		Name: "Mateo",
	})
	assert.NoError(t, err)
	assert.NotNil(t, user)
	t.Logf("User: %#v", user)

	users, err := client.Users(url.Values{"id_external": {user.ID}})
	assert.NoError(t, err)
	assert.NotNil(t, users)
}

func TestCreateEmptySession(t *testing.T) {
	session, err := client.CreateSession(&User{})
	assert.Error(t, err)
	assert.Nil(t, session)
}

func TestCreateSession(t *testing.T) {
	user, err := client.CreateUser(&User{
		Name: "Mateo",
	})
	assert.NoError(t, err)
	assert.NotNil(t, user)

	session, err := client.CreateSession(user)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	t.Logf("Session: %#v", session)

	valid, err := client.ValidToken(session.Token)
	assert.NoError(t, err)
	assert.True(t, valid)

	valid, err = client.ValidToken("WAKA")
	assert.Error(t, err)
	assert.False(t, valid)

	removed, err := client.RemoveToken(session.Token)
	assert.NoError(t, err)
	assert.True(t, removed)
}

func TestGetSiteOrganizations(t *testing.T) {
	sites, err := client.SiteOrganizations()
	assert.NoError(t, err)
	assert.NotNil(t, sites)
	t.Logf("Sites: %v", sites)
}

func TestGetCatalogues(t *testing.T) {
	catalogues, err := client.Catalogues(nil)
	assert.NoError(t, err)
	assert.NotNil(t, catalogues)
	t.Logf("Catalogues: %v", catalogues)
}

func TestCredentials(t *testing.T) {
	user, err := client.CreateUser(&User{
		Name: "Mateo",
	})
	assert.NoError(t, err)
	assert.NotNil(t, user)

	catalogues, err := client.Catalogues(url.Values{"is_test": {"true"}})
	assert.NoError(t, err)
	assert.NotNil(t, catalogues)

	for _, catalogue := range catalogues {
		session, err := client.CreateSession(user)
		assert.NoError(t, err)
		assert.NotNil(t, session)

		credentials := map[string]string{}
		for _, cred := range catalogue.Credentials {
			if cred.Required {
				credentials[cred.Name] = "test"
			}
		}

		credential, err := client.CreateCredential(&CredentialRequest{
			IDSite:      catalogue.IDSite,
			Credentials: credentials,
			Token:       session.Token,
		})
		assert.NoError(t, err)
		assert.NotNil(t, credential)

		t.Logf("New credential: %#v", credential)

		time.Sleep(time.Second * 5)
		status, err := client.Status(credential.Status, url.Values{"token": {session.Token}})
		assert.NoError(t, err)
		t.Logf("New status: %#v", status)

		if status.Last() == 200 {
			accounts, err := client.Accounts(url.Values{"token": {session.Token}})
			assert.NoError(t, err)

			t.Logf("accounts: %#v", accounts)

			transactions, err := client.Transactions(url.Values{"token": {session.Token}})
			assert.NoError(t, err)

			t.Logf("transactions: %#v", transactions)
		}

		assert.NotEqual(t, 0, len(status))
	}
}
