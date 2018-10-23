package rancher2

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	"github.com/rancher/types/client/management/v3"
)

// clusterByName returns the configuration of the cluster with the given name if it exists.
// Cluster names must always be unique, otherwise Rancher will yell at us...
func clusterByName(c *client.Client, name string) (*client.Cluster, error) {
	clusters, err := c.Cluster.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"name": name,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(clusters.Data) > 0 {
		return &clusters.Data[0], nil
	}
	return nil, nil
}

func resourceClusterCreate(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	name := d.Get("name").(string)

	if cluster, err := clusterByName(rancher, name); err != nil {
		return err
	} else if cluster != nil {
		return fmt.Errorf("cluster with name \"%s\" (ID: \"%s\") does already exist", name, cluster.ID)
	}

	newCluster, err := rancher.Cluster.Create(&client.Cluster{
		Name:        name,
		Description: d.Get("description").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(newCluster.ID)
	d.Set("cluster_id", newCluster.ID)
	d.Set("uuid", newCluster.UUID)

	return nil
}

func resourceClusterRead(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	if cluster, err := rancher.Cluster.ByID(d.Id()); err != nil {
		return err
	} else if cluster == nil {
		// If the cluster DOES NOT EXIST, it has probably already been deleted. Time to update the state...
		d.SetId("")
		return nil
	} else {
		if err := d.Set("cluster_id", cluster.ID); err != nil {
			return err
		}
		if err := d.Set("name", cluster.Name); err != nil {
			return err
		}
		if err := d.Set("description", cluster.Description); err != nil {
			return err
		}
		if err := d.Set("uuid", cluster.UUID); err != nil {
			return err
		}
	}
	return nil
}

func resourceClusterUpdate(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	id := d.Id()

	d.Partial(true)

	cluster, err := rancher.Cluster.ByID(id)
	if err != nil {
		return err
	}
	if cluster == nil {
		return fmt.Errorf("cluster with ID \"%s\" could not be found", id)
	}

	name := d.Get("name").(string)

	updates := map[string]string{
		// The name has to be given every time, even if it doesn't change, otherwise Rancher will complain.
		"name": name,
	}

	if d.HasChange("name") {
		// It is possible (even if it is unlikely!), that someone has already changed the name of the
		// cluster by other means, e.g. clicking on the UI. If the name has already been changed what
		// we want to apply via terraform, we'll just apply partial state and won't check for clusters
		// with duplicate names.
		if name == cluster.Name {
			d.SetPartial("name")
		} else {
			if clusterWithSameName, err := clusterByName(rancher, name); err != nil {
				return err
			} else if clusterWithSameName != nil {
				return fmt.Errorf("cluster with name \"%s\" (ID: \"%s\") does already exist", name, clusterWithSameName.ID)
			}
		}
	}

	if d.HasChange("description") {
		updates["description"] = d.Get("description").(string)
	}

	if _, err = rancher.Cluster.Update(cluster, updates); err != nil {
		return err
	}

	d.Partial(false)

	return nil

}

func resourceClusterDelete(d *schema.ResourceData, m interface{}) error {
	rancher := m.(Config).Rancher()

	id := d.Id()
	cluster, err := rancher.Cluster.ByID(id)
	if err != nil {
		return err
	}
	if cluster == nil {
		// If the cluster DOES NOT EXIST, it has probably already been deleted. Nothing to do for us here...
		return nil
	}
	return rancher.Cluster.Delete(cluster)
}

func resourceClusterExists(d *schema.ResourceData, m interface{}) (bool, error) {
	rancher := m.(Config).Rancher()
	id := d.Id()
	cluster, err := rancher.Cluster.ByID(id)
	if err != nil {
		if _, isApiError := err.(*clientbase.APIError); isApiError && err.(*clientbase.APIError).StatusCode == 404 {
			// Ignore 404 errors...
			return false, nil
		}
	}
	return cluster != nil, err
}

func resourceClusterState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := resourceClusterRead(d, m); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceClusterCreate,
		Read:   resourceClusterRead,
		Update: resourceClusterUpdate,
		Delete: resourceClusterDelete,
		Exists: resourceClusterExists,
		Importer: &schema.ResourceImporter{
			State: resourceClusterState,
		},
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Description: "Cluster ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name of the new cluster",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description to apply to the cluster",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"uuid": {
				Description: "UUID of the cluster as reported by the Rancher API",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}
