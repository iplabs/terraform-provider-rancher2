package rancher2

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	"github.com/rancher/types/client/management/v3"
)

// userByUsername returns the user with the given username/login if it exists.
// Users' usernames must always be unique for a given cluster, otherwise Rancher will yell at us...
func userByUsername(c *client.Client, username string) (*client.User, error) {
	users, err := c.User.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"username": username,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(users.Data) > 0 {
		return &users.Data[0], nil
	}
	return nil, nil
}

func getActiveDirectoryUser(u *client.User) string {
	pfx := "activedirectory_user://"
	pfxLen := len(pfx)
	for _, id := range u.PrincipalIDs {
		idLen := len(id)
		if strings.Index(id, pfx) == 0 {
			return id[pfxLen : idLen-pfxLen]
		}
	}
	return ""
}

func newUserPassword(d *schema.ResourceData) (password string) {
	if password = d.Get("password").(string); password == "" {
		password = fmt.Sprintf("%d%d", rand.New(rand.NewSource(time.Now().UnixNano())).Int(), rand.New(rand.NewSource(time.Now().UnixNano())).Int())
	}
	return password
}

func resourceUserCreate(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	username := d.Get("username").(string)

	password := ""
	if username != "" {
		if user, err := userByUsername(rancher, username); err != nil {
			return err
		} else if user != nil {
			return fmt.Errorf("user with username \"%s\" (ID: \"%s\") does already exist", username, user.ID)
		}
		password = newUserPassword(d)
	}

	principalIDs := make([]string, 0)
	activeDirectoryUser := d.Get("activedirectory_user").(string)
	if activeDirectoryUser != "" {
		principalIDs = append(principalIDs, fmt.Sprintf("activedirectory_user://%s", activeDirectoryUser))
	}

	user, err := rancher.User.Create(&client.User{
		Username:           username,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		PrincipalIDs:       principalIDs,
		Password:           password,
		MustChangePassword: false,
	})
	if err != nil {
		return err
	}

	d.SetId(user.ID)
	d.Set("user_id", user.ID)
	d.Set("uuid", user.UUID)
	d.Set("name", user.Name)
	d.Set("password", password)
	d.Set("description", user.Description)
	d.Set("activedirectory_user", getActiveDirectoryUser(user))

	return nil
}

func resourceUserRead(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	user, err := rancher.User.ByID(d.Id())

	if err != nil {
		return err
	} else if user == nil {
		// If the user DOES NOT EXIST, he/she has probably already been deleted. Time to update the state...
		d.SetId("")
		return nil
	}
	d.Set("user_id", user.ID)
	d.Set("uuid", user.UUID)
	d.Set("username", user.Username)
	d.Set("name", user.Name)
	d.Set("description", user.Description)
	d.Set("activedirectory_user", getActiveDirectoryUser(user))
	return nil
}

func resourceUserUpdate(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	id := d.Id()

	d.Partial(true)

	user, err := rancher.User.ByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user with ID \"%s\" could not be found", id)
	}

	updates := map[string]interface{}{}
	if d.HasChange("username") {
		updates["username"] = d.Get("username").(string)
	}
	if d.HasChange("password") {
		updates["password"] = d.Get("password").(string)
	}
	if d.HasChange("name") {
		updates["name"] = d.Get("name").(string)
	}
	if d.HasChange("description") {
		updates["description"] = d.Get("description").(string)
	}
	if d.HasChange("activedirectory_user") {
		updates["principalIDs"] = []string{
			fmt.Sprintf("activedirectory_user://%s", d.Get("activedirectory_user").(string)),
		}
	}

	if _, err = rancher.User.Update(user, updates); err != nil {
		return err
	}

	d.Partial(false)

	return nil

}

func resourceUserDelete(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	id := d.Id()
	user, err := rancher.User.ByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		// If the user DOES NOT EXIST, he/she has probably already been deleted. Nothing to do for us here...
		return nil
	}
	return rancher.User.Delete(user)
}

func resourceUserExists(d *schema.ResourceData, m interface{}) (bool, error) {
	rancher := m.(Config).Rancher()

	user, err := rancher.User.ByID(d.Id())
	if err != nil {
		if _, isAPIError := err.(*clientbase.APIError); isAPIError && err.(*clientbase.APIError).StatusCode == 404 {
			// Ignore 404 errors...
			return false, nil
		}
	}
	return user != nil, nil
}

func resourceUserState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := resourceUserRead(d, m); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserCreate,
		Read:   resourceUserRead,
		Update: resourceUserUpdate,
		Delete: resourceUserDelete,
		Exists: resourceUserExists,
		Importer: &schema.ResourceImporter{
			State: resourceUserState,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Description: "ID of the Rancher user",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"uuid": {
				Description: "UUID of the Rancher user",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"username": {
				Description: "Username (login) of the Rancher user",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"password": {
				Description: "Password of the Rancher user",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old != "" && new == ""
				},
			},
			"name": {
				Description: "Display name of the Rancher user",
				Type:        schema.TypeString,
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Active directory user names seem to be being pre-populated
					return d.Get("activedirectory_user") != "" && new == ""
				},
			},
			"description": {
				Description: "Description of the Rancher user",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"activedirectory_user": {
				Description: "Distinguished name (DN) of the associated ActiveDirectory user account",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}
