package rancher2

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/rancher/norman/clientbase"
	"github.com/rancher/types/client/management/v3"
)

func resourceClusterRegistrationTokenCreate(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	newToken, err := rancher.ClusterRegistrationToken.Create(&client.ClusterRegistrationToken{
		ClusterID: d.Get("cluster_id").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(newToken.ID)
	d.Set("token", newToken.Token)
	d.Set("name", newToken.Name)
	d.Set("command", newToken.Command)
	d.Set("insecure_command", newToken.InsecureCommand)
	d.Set("manifest_url", newToken.ManifestURL)

	return nil
}

func resourceClusterRegistrationTokenRead(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	token, err := rancher.ClusterRegistrationToken.ByID(d.Id())

	if err != nil {
		return err
	} else if token == nil {
		// If the token DOES NOT EXIST, it has probably already been deleted. Time to update the state...
		d.SetId("")
		return nil
	}
	d.Set("token", token.Token)
	d.Set("name", token.Name)
	d.Set("command", token.Command)
	d.Set("insecure_command", token.InsecureCommand)
	d.Set("manifest_url", token.ManifestURL)
	return nil
}

func resourceClusterRegistrationTokenUpdate(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	id := d.Id()

	d.Partial(true)

	token, err := rancher.ClusterRegistrationToken.ByID(id)
	if err != nil {
		return err
	}
	if token == nil {
		return fmt.Errorf("registration token with ID \"%s\" could not be found", id)
	}

	updates := map[string]string{}
	if d.HasChange("name") {
		updates["name"] = d.Get("name").(string)
	}

	if _, err = rancher.ClusterRegistrationToken.Update(token, updates); err != nil {
		return err
	}

	d.Partial(false)

	return nil

}

func resourceClusterRegistrationTokenDelete(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	id := d.Id()
	token, err := rancher.ClusterRegistrationToken.ByID(id)
	if err != nil {
		return err
	}
	if token == nil {
		// If the token DOES NOT EXIST, it has probably already been deleted. Nothing to do for us here...
		return nil
	}
	return rancher.ClusterRegistrationToken.Delete(token)
}

func resourceClusterRegistrationTokenExists(d *schema.ResourceData, m interface{}) (bool, error) {
	rancher := m.(Config).Rancher()

	token, err := rancher.ClusterRegistrationToken.ByID(d.Id())
	if err != nil {
		if _, isApiError := err.(*clientbase.APIError); isApiError && err.(*clientbase.APIError).StatusCode == 404 {
			// Ignore 404 errors...
			return false, nil
		}
	}
	return token != nil, nil
}

func resourceClusterRegistrationToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceClusterRegistrationTokenCreate,
		Read:   resourceClusterRegistrationTokenRead,
		Update: resourceClusterRegistrationTokenUpdate,
		Delete: resourceClusterRegistrationTokenDelete,
		Exists: resourceClusterRegistrationTokenExists,

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Description: "ID of the cluster for whom to create the token",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Human-readable name for the token",
				Type:        schema.TypeString,
				Optional:    true,
				DiffSuppressFunc: func(k string, old string, new string, d *schema.ResourceData) bool {
					// If the token has given a value that has been computed by Rancher, we ignore
					// the "blank string default input" of optional string fields.
					return new == ""
				},
			},
			"token": {
				Description: "Registration token string",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"manifest_url": {
				Description: "URL to the kubernetes manifest that includes the existing cluster into Rancher",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"command": {
				Description: "Command to include the existing cluster into Rancher",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"insecure_command": {
				Description: "Insecure command to include the existing cluster into Rancher",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}
