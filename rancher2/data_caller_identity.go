package rancher2

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
)

func dataCallerIdentityRead(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	filters := map[string]interface{}{
		"me":      true,
		"enabled": true,
	}

	users, err := rancher.User.List(&types.ListOpts{
		Filters: filters,
	})
	if err != nil {
		return err
	}
	if len(users.Data) <= 0 {
		return fmt.Errorf("could not determine current Rancher user")
	}
	if len(users.Data) > 1 {
		return fmt.Errorf("more than one current Rancher user found (this should never happen!)")
	}

	user := users.Data[0] // We simply use the first token that matches the search criteria.
	d.SetId(user.ID)
	d.Set("name", user.Name)
	d.Set("description", user.Description)
	d.Set("uuid", user.UUID)
	return nil
}

func dataCallerIdentityExists(d *schema.ResourceData, m interface{}) (bool, error) {
	rancher := m.(Config).Rancher()

	id := d.Id()
	user, err := rancher.User.ByID(id)
	if err != nil {
		if _, isApiError := err.(*clientbase.APIError); isApiError && err.(*clientbase.APIError).StatusCode == 404 {
			// Ignore 404 errors...
			return false, nil
		}
	}
	return user != nil, err
}

func dataCallerIdentity() *schema.Resource {
	return &schema.Resource{
		Read:   dataCallerIdentityRead,
		Exists: dataCallerIdentityExists,
		Schema: map[string]*schema.Schema{
			"name": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"description": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"uuid": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}
