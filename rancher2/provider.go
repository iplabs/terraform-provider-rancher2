package rancher2

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

type RancherCLIConfiguration struct {
	Servers map[string]struct {
		AccessKey string `json:"accessKey"`
		SecretKey string `json:"secretKey"`
		CACert    string `json:"cacert"`
		URL       string `json:"url"`
	} `json:"Servers"`
	CurrentServer string `json:"CurrentServer"`
}

// Provider creates a new Rancher 2 terraform provider instance.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_url": {
				Description: "URL of the rancher server instance.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"access_key": {
				Description: "Rancher API access key",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"secret_key": {
				Description: "Rancher API secret key",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"token": {
				Description: "Rancher API access token (can be used if neither access_key or secret_key) have been provided",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"cacert": {
				Description: "Rancher CA certificate",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"current_server": {
				Description: "Current rancher server configuration (if using Rancher CLI settings)",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"rancher2_caller_identity": dataCallerIdentity(),
			"rancher2_project":         dataProject(),
			"rancher2_token":           dataToken(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"rancher2_cluster":                    resourceCluster(),
			"rancher2_cluster_registration_token": resourceClusterRegistrationToken(),
			"rancher2_project":                    resourceProject(),
			"rancher2_token":                      resourceToken(),
			"rancher2_user":                       resourceUser(),
		},
		ConfigureFunc: configure,
	}
}

func readCLIConfiguration(r io.Reader) (*RancherCLIConfiguration, error) {
	cfg := RancherCLIConfiguration{}
	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func configure(d *schema.ResourceData) (interface{}, error) {
	apiURL := d.Get("api_url").(string)
	accessKey := d.Get("access_key").(string)
	secretKey := d.Get("secret_key").(string)
	token := d.Get("token").(string)
	if token != "" {
		keys := strings.Split(token, ":")
		accessKey = keys[0]
		secretKey = keys[1]
	}
	cacert := d.Get("cacert").(string)
	currentServer := d.Get("current_server").(string)

	// Let's try to read the rancher CLI configuration file if the user did not specify
	// any connection parameters in the provider configuration.
	if apiURL == "" && token == "" && accessKey == "" && secretKey == "" && cacert == "" {
		cu, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("unable to determine current user: %v", err)
		}
		cfgFilePath := fmt.Sprintf("%s/.rancher/cli2.json", cu.HomeDir)
		if _, err := os.Stat(cfgFilePath); err != nil {
			return nil, fmt.Errorf("unable to find Rancher CLI configuration file %s: %v", cfgFilePath, err)
		}
		cfgFile, err := os.Open(cfgFilePath)
		if err != nil {
			return nil, fmt.Errorf("unable to open Rancher CLI configuration file %s: %v", cfgFile.Name(), err)
		}
		defer cfgFile.Close()
		cfg, err := readCLIConfiguration(cfgFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read Rancher CLI configuration file %s: %v", cfgFile.Name(), err)
		}
		if currentServer == "" {
			currentServer = cfg.CurrentServer
		}
		server, serverExists := cfg.Servers[currentServer]
		if !serverExists {
			return nil, fmt.Errorf("unable to find server with ID '%s' in Rancher CLI configuration file %s", currentServer, cfgFile.Name())
		}
		apiURL = server.URL
		accessKey = server.AccessKey
		secretKey = server.SecretKey
		cacert = server.CACert
	} else {
		cfgMissing := make([]string, 0)
		if apiURL == "" {
			cfgMissing = append(cfgMissing, "api_url")
		}
		if accessKey == "" {
			cfgMissing = append(cfgMissing, "access_key")
		}
		if secretKey == "" {
			cfgMissing = append(cfgMissing, "secret_key")
		}
		if len(cfgMissing) > 0 {
			return nil, fmt.Errorf("required configuration parameter(s): %s", strings.Join(cfgMissing, ","))
		}
	}

	return NewConfig(apiURL, accessKey, secretKey, cacert)
}
