package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccServerReverseResourceConfig(name, family, reverse string) string {
	return fmt.Sprintf(`
resource "gigahost_server" "test" {
  product_name = %q
  region_name  = "Sandefjord"
  rescue       = true
  srv_name     = %q
  timeouts     = { create = "15m" }
}

locals {
  primary_ip = one([for ip in gigahost_server.test.ips : ip if ip.ip_v4v6 == %q && ip.ip_type == "primary"])
}

resource "gigahost_server_reverse" "test" {
  srv_id     = gigahost_server.test.srv_id
  ip_id      = local.primary_ip.ip_id
  ip_v4v6    = local.primary_ip.ip_v4v6
  ip_reverse = %q
}
`, testAccServerProduct(), name, family, reverse)
}

func TestAccServerReverseResource_primaryIPv4(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerReverseResourceConfig("tf-acc-server-reverse-v4", "ipv4", "tf-acc-v4.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("gigahost_server_reverse.test", "ip_address"),
					resource.TestCheckResourceAttr("gigahost_server_reverse.test", "ip_type", "primary"),
					resource.TestCheckResourceAttr("gigahost_server_reverse.test", "ip_v4v6", "ipv4"),
					resource.TestCheckResourceAttr("gigahost_server_reverse.test", "ip_reverse", "tf-acc-v4.example.com"),
				),
			},
			{
				ResourceName:      "gigahost_server_reverse.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccServerReverseImportID,
			},
			{
				Config: testAccServerReverseResourceConfig("tf-acc-server-reverse-v4", "ipv4", ""),
				Check:  resource.TestCheckResourceAttr("gigahost_server_reverse.test", "ip_reverse", ""),
			},
			{
				ResourceName:      "gigahost_server_reverse.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccServerReverseImportID,
			},
		},
	})
}

func TestAccServerReverseResource_primaryIPv6(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerReverseResourceConfig("tf-acc-server-reverse-v6", "ipv6", "tf-acc-v6.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("gigahost_server_reverse.test", "ip_address"),
					resource.TestCheckResourceAttr("gigahost_server_reverse.test", "ip_type", "primary"),
					resource.TestCheckResourceAttr("gigahost_server_reverse.test", "ip_v4v6", "ipv6"),
					resource.TestCheckResourceAttr("gigahost_server_reverse.test", "ip_reverse", "tf-acc-v6.example.com"),
				),
			},
			{
				ResourceName:            "gigahost_server_reverse.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccServerReverseImportID,
				ImportStateVerifyIgnore: []string{"ip_reverse"},
			},
			{
				Config: testAccServerReverseResourceConfig("tf-acc-server-reverse-v6", "ipv6", ""),
				Check:  resource.TestCheckResourceAttr("gigahost_server_reverse.test", "ip_reverse", ""),
			},
			{
				ResourceName:      "gigahost_server_reverse.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccServerReverseImportID,
			},
		},
	})
}

func testAccServerReverseImportID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["gigahost_server_reverse.test"]
	if !ok {
		return "", fmt.Errorf("resource not found in state")
	}
	attrs := rs.Primary.Attributes
	return fmt.Sprintf("%s/%s/%s", attrs["srv_id"], attrs["ip_v4v6"], attrs["ip_id"]), nil
}
