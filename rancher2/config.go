package rancher2

import (
	"fmt"
	"github.com/rancher/norman/clientbase"
	rancher "github.com/rancher/types/client/management/v3"
	"strings"
)

// Config provides a way to wrap/encapsulate all the stuff that is necessary
// to interact with the Rancher REST API.
type Config interface {
	Rancher() *rancher.Client
}

type config struct {
	url           string
	cacert        string
	accessKey     string
	secretKey     string
	rancherClient *rancher.Client
}

func (c *config) Rancher() *rancher.Client {
	return c.rancherClient
}

// NewConfig creates a new configuration structure to be used provider-internally.
func NewConfig(url string, accessKey string, secretKey string, cacert string) (Config, error) {

	// Normalize URL configuration. We want to forgive user inputs...
	if strings.HasSuffix(url, "/") {
		url = url[0 : len(url)-1] // Chop off trailing slashes! (Rancher API server doesn't like them!)
	}
	if !strings.HasSuffix(url, "/v3") {
		url = fmt.Sprintf("%s/v3", url) // We use API v3 to communicate with Rancher 2 instances.
	}

	// Creates a new Rancher client instance that shall be used for all sub-sequent
	// calls to the API.
	rancherClient, err := rancher.NewClient(&clientbase.ClientOpts{
		URL:       url,
		AccessKey: accessKey,
		SecretKey: secretKey,
		CACerts:   cacert,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create Rancher client: %v", err)
	}

	return &config{
		url:           url,
		cacert:        cacert,
		accessKey:     accessKey,
		secretKey:     secretKey,
		rancherClient: rancherClient,
	}, nil

}
