package rancher2

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/rancher/norman/clientbase"
	"github.com/rancher/types/client/management/v3"
	"strings"
)

func resourceTokenCreate(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	token, err := rancher.Token.Create(&client.Token{
		Description: d.Get("description").(string),
		UserID:      d.Get("user_id").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(token.ID)
	d.Set("user_id", token.UserID)
	d.Set("description", token.Description)
	d.Set("token", token.Token)

	key := strings.Split(token.Token, ":")
	d.Set("access_key", key[0])
	d.Set("secret_key", key[1])

	return nil
}

func resourceTokenRead(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	token, err := rancher.Token.ByID(d.Id())

	if err != nil {
		return err
	} else if token == nil {
		// If the token DOES NOT EXIST, it has probably already been deleted. Time to update the state...
		d.SetId("")
		return nil
	}
	d.Set("user_id", token.UserID)
	d.Set("description", token.Description)
	return nil
}

func resourceTokenUpdate(d *schema.ResourceData, m interface{}) error {
	// TODO Implement me?
	return nil

}

func resourceTokenDelete(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	id := d.Id()
	token, err := rancher.Token.ByID(id)
	if err != nil {
		return err
	}
	if token == nil {
		// If the token DOES NOT EXIST, it has probably already been deleted. Nothing to do for us here...
		return nil
	}
	return rancher.Token.Delete(token)
}

func resourceTokenExists(d *schema.ResourceData, m interface{}) (bool, error) {
	rancher := m.(Config).Rancher()

	token, err := rancher.Token.ByID(d.Id())
	if err != nil {
		if _, isAPIError := err.(*clientbase.APIError); isAPIError && err.(*clientbase.APIError).StatusCode == 404 {
			// Ignore 404 errors...
			return false, nil
		}
	}
	return token != nil, nil
}

func resourceTokenState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := resourceTokenRead(d, m); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceTokenCreate,
		Read:   resourceTokenRead,
		Update: resourceTokenUpdate,
		Delete: resourceTokenDelete,
		Exists: resourceTokenExists,
		Importer: &schema.ResourceImporter{
			State: resourceTokenState,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Description: "ID of the user for whom to create the token",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description of the token",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"token": {
				Description: "API token",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"access_key": {
				Description: "Access key",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"secret_key": {
				Description: "Secret key",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}
