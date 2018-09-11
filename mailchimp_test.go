package mailchimp

import (
	"context"
	"testing"
)

func TestNewClientWithValidApiKey(t *testing.T) {
	_, err := NewChimpApi("myapikey-us20")
	if err != nil {
		t.Errorf("error parsing api key: %v", err)
	}
}
func TestNewClientWithInValidApiKey(t *testing.T) {
	_, err := NewChimpApi("myapikey")
	if err == nil {
		t.Errorf("invalid key format")
	}
}

func TestNewClientHasValidBaseUrl(t *testing.T) {
	api, err := NewChimpApi("myapikey-us20")
	if err != nil {
		t.Errorf("invalid key format %v", err)
	}
	c := api.GetClient(context.Background())
	url := c.(*client).baseUrl.String()
	expected := "https://us20.api.mailchimp.com/3.0/"

	if url != expected {
		t.Errorf("parsed base url incorrected baseed on the api key, expected %s, got %s", url, expected)
	}
}
