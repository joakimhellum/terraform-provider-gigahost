package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccServerIPv4Resource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerIPv4ResourceConfig("tf-acc-ipv4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gigahost_server_ipv4.l2", "ip_type", "l2"),
					resource.TestCheckResourceAttrSet("gigahost_server_ipv4.l2", "ip_id"),
					resource.TestCheckResourceAttrSet("gigahost_server_ipv4.l2", "ip_address"),
					resource.TestCheckResourceAttr("gigahost_server_ipv4.l3", "ip_type", "l3"),
					resource.TestCheckResourceAttrSet("gigahost_server_ipv4.l3", "ip_id"),
					resource.TestCheckResourceAttrSet("gigahost_server_ipv4.l3", "ip_address"),
				),
			},
			{
				ResourceName:      "gigahost_server_ipv4.l2",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccServerIPv4ImportID("gigahost_server_ipv4.l2"),
			},
			{
				ResourceName:      "gigahost_server_ipv4.l3",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccServerIPv4ImportID("gigahost_server_ipv4.l3"),
			},
		},
	})
}

func testAccServerIPv4ResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "gigahost_server" "test" {
  product_name = %q
  region_name  = "Sandefjord"
  rescue       = true
  srv_name     = %q
  timeouts     = { create = "15m" }
}

resource "gigahost_server_ipv4" "l2" {
  srv_id  = gigahost_server.test.srv_id
  ip_type = "l2"
}

resource "gigahost_server_ipv4" "l3" {
  srv_id  = gigahost_server.test.srv_id
  ip_type = "l3"
}
`, testAccServerProduct(), name)
}

func testAccServerIPv4ImportID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found in state", resourceName)
		}
		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["srv_id"], rs.Primary.Attributes["ip_id"], rs.Primary.Attributes["ip_type"]), nil
	}
}
