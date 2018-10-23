package rancher2

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	"github.com/rancher/types/client/management/v3"
)

// projectByName returns the configuration of the project with the given name if it exists.
// Project names must always be unique for a given cluster, otherwise Rancher will yell at us...
func projectByClusterAndName(c *client.Client, clusterID string, name string) (*client.Project, error) {
	projects, err := c.Project.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"clusterId": clusterID,
			"name":      name,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(projects.Data) > 0 {
		return &projects.Data[0], nil
	}
	return nil, nil
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	clusterID := d.Get("cluster_id").(string)
	name := d.Get("name").(string)

	if project, err := projectByClusterAndName(rancher, clusterID, name); err != nil {
		return err
	} else if project != nil {
		return fmt.Errorf("project with name \"%s\" (ID: \"%s\") does already exist", name, project.ID)
	}

	project, err := rancher.Project.Create(&client.Project{
		Name:        name,
		Description: d.Get("description").(string),
		ClusterID:   clusterID,
	})
	if err != nil {
		return err
	}

	d.SetId(project.ID)
	d.Set("cluster_id", project.ClusterID)
	d.Set("name", project.Name)
	d.Set("description", project.Description)

	return nil
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	project, err := rancher.Project.ByID(d.Id())

	if err != nil {
		return err
	} else if project == nil {
		// If the project DOES NOT EXIST, it has probably already been deleted. Time to update the state...
		d.SetId("")
		return nil
	}
	d.SetId(project.ID)
	d.Set("cluster_id", project.ClusterID)
	d.Set("name", project.Name)
	d.Set("description", project.Description)
	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	id := d.Id()

	d.Partial(true)

	project, err := rancher.Project.ByID(id)
	if err != nil {
		return err
	}
	if project == nil {
		return fmt.Errorf("project with ID \"%s\" could not be found", id)
	}

	updates := map[string]string{}
	if d.HasChange("name") {
		updates["name"] = d.Get("name").(string)
	}
	if d.HasChange("description") {
		updates["description"] = d.Get("description").(string)
	}

	if _, err = rancher.Project.Update(project, updates); err != nil {
		return err
	}

	d.Partial(false)

	return nil

}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	id := d.Id()
	project, err := rancher.Project.ByID(id)
	if err != nil {
		return err
	}
	if project == nil {
		// If the project DOES NOT EXIST, it has probably already been deleted. Nothing to do for us here...
		return nil
	}
	return rancher.Project.Delete(project)
}

func resourceProjectExists(d *schema.ResourceData, m interface{}) (bool, error) {
	rancher := m.(Config).Rancher()

	project, err := rancher.Project.ByID(d.Id())
	if err != nil {
		if _, isApiError := err.(*clientbase.APIError); isApiError && err.(*clientbase.APIError).StatusCode == 404 {
			// Ignore 404 errors...
			return false, nil
		}
	}
	return project != nil, nil
}

func resourceProjectState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := resourceProjectRead(d, m); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,
		Exists: resourceProjectExists,
		Importer: &schema.ResourceImporter{
			State: resourceProjectState,
		},
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Description: "ID of the cluster for whom to create the project",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Name of the project",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description of the project",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}
