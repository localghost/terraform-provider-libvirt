package libvirt

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider libvirt.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"uri": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
                ConflictsWith: []string{"uris"},
				Description: "libvirt connection URI for operations. See https://libvirt.org/uri.html",
			},
			"uris": {
				Type:     schema.TypeList,
				Required: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
                ConflictsWith: []string{"uri"},
				Description: "libvirt connection URIs for operations. See https://libvirt.org/uri.html",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"libvirt_domain":         resourceLibvirtDomain(),
			"libvirt_volume":         resourceLibvirtVolume(),
			"libvirt_network":        resourceLibvirtNetwork(),
			"libvirt_pool":           resourceLibvirtPool(),
			"libvirt_cloudinit_disk": resourceCloudInitDisk(),
			"libvirt_ignition":       resourceIgnition(),
			"libvirt_combustion":     resourceCombustion(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"libvirt_network_dns_host_template":        datasourceLibvirtNetworkDNSHostTemplate(),
			"libvirt_network_dns_srv_template":         datasourceLibvirtNetworkDNSSRVTemplate(),
			"libvirt_network_dnsmasq_options_template": datasourceLibvirtNetworkDnsmasqOptionsTemplate(),
			"libvirt_node_info":                        datasourceLibvirtNodeInfo(),
			"libvirt_node_device_info":                 datasourceLibvirtNodeDeviceInfo(),
			"libvirt_node_devices":                     datasourceLibvirtNodeDevices(),
		},

		ConfigureFunc: providerConfigure,
	}
}

// uri -> client for multi instance support
// (we share the same client for the same uri).
var globalClientMap = make(map[string]*Client)

// CleanupLibvirtConnections closes libvirt clients for all URIs.
func CleanupLibvirtConnections() {
	for uri, client := range globalClientMap {
		log.Printf("[DEBUG] cleaning up connection for URI: %s", uri)
		// TODO: Confirm appropriate IsAlive() validation
		err := client.libvirt.ConnectClose()
		if err != nil {
			log.Printf("[ERROR] cannot close libvirt connection: %v", err)
		}
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	var libvirt_uris []interface{}
	if uri, ok := d.GetOk("uri"); ok {
		libvirt_uris = append(libvirt_uris, uri)
	} else if uris, ok := d.GetOk("uris"); ok {
        libvirt_uris = uris.([]interface{})
	} else {
        uri := os.Getenv("LIBVIRT_DEFAULT_URI")
        if uri != "" {
            libvirt_uris = append(libvirt_uris, uri)
        }
    }
    if len(libvirt_uris) == 0 {
		return nil, fmt.Errorf("neither uri, uris nor LIBVIRT_DEFAULT_URI environment variable set")
    }

	result := make(map[string]*Client)
	for _, uri := range libvirt_uris {
        uri := uri.(string)
		config := Config{
			URI: uri,
		}
		log.Printf("[DEBUG] Configuring provider for '%s': %v", config.URI, d)

		if client, ok := globalClientMap[config.URI]; ok {
			log.Printf("[DEBUG] Reusing client for uri: '%s'", config.URI)
			result[config.URI] = client
		} else {
			client, err := config.Client()
			if err != nil {
				return nil, err
			}
			globalClientMap[config.URI] = client
			result[config.URI] = client
		}
	}
    result["default"] = result[libvirt_uris[0].(string)]
	return result, nil
}
