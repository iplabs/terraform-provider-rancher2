package rancher2

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/rancher/norman/types"
)

func dataProjectRead(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	clusterID := d.Get("cluster_id").(string)
	name := d.Get("name").(string)

	projects, err := rancher.Project.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"clusterId": clusterID,
			"name":      name,
		},
	})
	if err != nil {
		return err
	}

	cnt := len(projects.Data)
	if cnt <= 0 {
		return fmt.Errorf("project with name \"%s\" not found", name)
	}
	if cnt > 1 {
		return fmt.Errorf("more than one project with specified name (\"%s\") found: %d", name, cnt)
	}

	// Only one project returned? Great...
	project := projects.Data[0]
	d.SetId(project.ID)
	d.Set("clusterId", project.ClusterID)
	d.Set("uuid", project.UUID)
	d.Set("name", project.Name)
	d.Set("description", project.Description)

	return nil
}

func dataProjectExists(d *schema.ResourceData, m interface{}) (bool, error) {
	rancher := m.(Config).Rancher()

	clusterID := d.Get("cluster_id").(string)
	name := d.Get("name").(string)

	projects, err := rancher.Project.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"cluster_id": clusterID,
			"name":       name,
		},
	})
	if err != nil {
		return false, err
	}

	cnt := len(projects.Data)
	if cnt <= 0 {
		return false, nil
	}
	if cnt > 1 {
		return false, fmt.Errorf("more than one project with specified name (\"%s\") found: %d", name, cnt)
	}

	// Only one project returned? Great...
	return true, nil
}

func dataProject() *schema.Resource {
	return &schema.Resource{
		Read:   dataProjectRead,
		Exists: dataProjectExists,
		Schema: map[string]*schema.Schema{
			"uuid": {
				Description: "UUID of the project",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cluster_id": {
				Description: "ID of the cluster for whom to create the project",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name of the project",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description of the project",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}
