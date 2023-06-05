package tencentcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTencentCloudVpcSubnetResourceDashboardDataSource_basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcSubnetResourceDashboardDataSource,
				Check:  resource.ComposeTestCheckFunc(testAccCheckTencentCloudDataSourceID("data.tencentcloud_vpc_subnet_resource_dashboard.subnet_resource_dashboard")),
			},
		},
	})
}

const testAccVpcSubnetResourceDashboardDataSource = `

data "tencentcloud_vpc_subnet_resource_dashboard" "subnet_resource_dashboard" {
  subnet_ids = ["subnet-i9tpf6hq"]
}

`