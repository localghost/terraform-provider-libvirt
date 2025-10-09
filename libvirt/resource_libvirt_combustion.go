package libvirt

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCombustion() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCombustionCreate,
		ReadContext:   resourceCombustionRead,
		DeleteContext: resourceCombustionDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pool": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
				ForceNew: true,
			},
			"content": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
            "libvirt_uri": {
                Type: schema.TypeString,
                Optional: true,
                ForceNew: true,
            },
		},
	}
}

func resourceCombustionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] creating combustion file")
    client := getLibvirtClientForResource(d, meta)

	combustion := newIgnitionDef()

	combustion.Name = d.Get("name").(string)
	combustion.PoolName = d.Get("pool").(string)
	combustion.Content = d.Get("content").(string)
	combustion.Combustion = true

	log.Printf("[INFO] combustion: %+v", combustion)

	key, err := combustion.CreateAndUpload(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(key)

	return resourceIgnitionRead(ctx, d, meta)
}

func resourceCombustionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    client := getLibvirtClientForResource(d, meta)
	virConn := client.libvirt

	combustion, err := newIgnitionDefFromRemoteVol(virConn, d.Id())
	d.Set("pool", combustion.PoolName)
	d.Set("name", combustion.Name)

	if err != nil {
		return diag.Errorf("error while retrieving remote volume: %s", err)
	}

	return nil
}

func resourceCombustionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    client := getLibvirtClientForResource(d, meta)

	key, err := getIgnitionVolumeKeyFromTerraformID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(volumeDelete(ctx, client, key))
}
