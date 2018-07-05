package opc

import (
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/client"
	"github.com/hashicorp/go-oracle-terraform/lbaas"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceLBaaSSSLCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceSSLCertificateCreate,
		Read:   resourceSSLCertificateRead,
		Delete: resourceSSLCertificateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateLoadBalancerResourceName,
			},
			"certificate_body": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// TODO ValidateFunc: PEM Key Format
			},
			"certificate_chain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// TODO ValidateFunc: PEM Key Format
			},
			"private_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				// TODO ValidateFunc:  PEM key format, only required when type = "SERVER"
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"SERVER",
					"TRUSTED",
				}, true),
			},

			// Read only attributes
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSSLCertificateCreate(d *schema.ResourceData, meta interface{}) error {

	sslCertClient := meta.(*Client).lbaasClient.SSLCertificateClient()

	input := lbaas.CreateSSLCertificateInput{
		Name:             d.Get("name").(string),
		Certificate:      d.Get("certificate_body").(string),
		CertificateChain: d.Get("certificate_chain").(string),
		PrivateKey:       d.Get("private_key").(string),
		Trusted:          d.Get("type").(string) == "TRUSTED",
	}

	info, err := sslCertClient.CreateSSLCertificate(&input)
	if err != nil {
		return fmt.Errorf("Error creating Load Balancer Server Pool: %s", err)
	}

	d.SetId(info.Name)
	return resourceSSLCertificateRead(d, meta)
}

func resourceSSLCertificateRead(d *schema.ResourceData, meta interface{}) error {
	sslCertClient := meta.(*Client).lbaasClient.SSLCertificateClient()
	name := d.Id()

	result, err := sslCertClient.GetSSLCertificate(name)
	if err != nil {
		// SSLCertificate does not exist
		if client.WasNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Server Pool %s: %s", d.Id(), err)
	}

	if result == nil {
		d.SetId("")
		return nil
	}

	d.Set("certificate_body", result.Certificate)
	d.Set("certificate_chain", result.CertificateChain)
	d.Set("name", result.Name)
	d.Set("state", result.State)
	d.Set("uri", result.URI)

	if result.Trusted {
		d.Set("type", "TRUSTED")
	} else {
		d.Set("type", "SERVER")
	}

	return nil
}

func resourceSSLCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	lbaasClient := meta.(*Client).lbaasClient.SSLCertificateClient()
	name := d.Id()

	if _, err := lbaasClient.DeleteSSLCertificate(name); err != nil {
		return fmt.Errorf("Error deleting SSLCertificate")
	}
	return nil
}