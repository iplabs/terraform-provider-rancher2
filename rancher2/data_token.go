package rancher2

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
)

func dataTokenRead(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	userID := d.Get("user_id").(string)

	filters := map[string]interface{}{
		"user_id": userID,
	}
	if expired, expiredSet := d.GetOkExists("expired"); expiredSet {
		filters["expired"] = expired
	}
	tokens, err := rancher.Token.List(&types.ListOpts{
		Filters: filters,
	})
	if err != nil {
		return err
	}
	if len(tokens.Data) <= 0 {
		return fmt.Errorf("no tokens for user %s have been found that match the search criteria", userID)
	}

	token := tokens.Data[0] // We simply use the first token that matches the search criteria.
	d.SetId(token.ID)
	d.Set("user_id", token.UserID)
	d.Set("is_derived", token.IsDerived)
	d.Set("expired", token.Expired)
	d.Set("name", token.Name)
	d.Set("description", token.Description)
	d.Set("uuid", token.UUID)
	return nil
}

func dataTokenExists(d *schema.ResourceData, m interface{}) (bool, error) {
	rancher := m.(Config).Rancher()

	id := d.Id()
	token, err := rancher.Token.ByID(id)
	if err != nil {
		if _, isApiError := err.(*clientbase.APIError); isApiError && err.(*clientbase.APIError).StatusCode == 404 {
			// Ignore 404 errors...
			return false, nil
		}
	}
	return token != nil, err
}

func dataToken() *schema.Resource {
	return &schema.Resource{
		Read:   dataTokenRead,
		Exists: dataTokenExists,
		Schema: map[string]*schema.Schema{
			"is_derived": {
				Computed: true,
				Type:     schema.TypeBool,
			},
			"user_id": {
				Required: true,
				Type:     schema.TypeString,
			},
			"expired": {
				Optional: true,
				Type:     schema.TypeBool,
			},
			"name": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"description": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"uuid": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}
