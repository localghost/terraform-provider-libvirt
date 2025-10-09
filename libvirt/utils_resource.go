package libvirt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getLibvirtClientForResource(d *schema.ResourceData, meta interface{}) *Client {
    var client *Client
    if uri, ok := d.GetOk("libvirt_uri"); ok {
        client = meta.(map[string]*Client)[uri.(string)]
    } else {
        client = meta.(map[string]*Client)["default"]
    }
    return client
}
