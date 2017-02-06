package paybook

import (
	"net/url"
	"os"
	"testing"

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
				credentials[cred.Name] = cred.Name
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

		status, err := client.Status(credential.Status, url.Values{"token": {session.Token}})
		assert.NoError(t, err)
		t.Logf("New status: %#v", status)

		assert.NotEqual(t, 0, len(status))
	}
}
