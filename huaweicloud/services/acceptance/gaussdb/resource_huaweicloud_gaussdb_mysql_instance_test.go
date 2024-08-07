package gaussdb

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/chnsz/golangsdk/openstack/taurusdb/v3/instances"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance/common"
)

func TestAccGaussDBInstance_basic(t *testing.T) {
	var instance instances.TaurusDBInstance

	name := acceptance.RandomAccResourceName()
	updateName := acceptance.RandomAccResourceName()
	resourceName := "huaweicloud_gaussdb_mysql_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      testAccCheckGaussDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGaussDBInstanceConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGaussDBInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "backup_strategy.0.start_time", "09:00-10:00"),
					resource.TestCheckResourceAttr(resourceName, "backup_strategy.0.keep_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "read_replicas", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "audit_log_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "sql_filter_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_id",
						"huaweicloud_gaussdb_mysql_parameter_template.test.0", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_name",
						"huaweicloud_gaussdb_mysql_parameter_template.test.0", "name"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.name", "default_authentication_plugin"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.value", "mysql_native_password"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
			{
				Config: testAccGaussDBInstanceConfig_basicUpdate(updateName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGaussDBInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "name", updateName),
					resource.TestCheckResourceAttr(resourceName, "backup_strategy.0.start_time", "12:00-13:00"),
					resource.TestCheckResourceAttr(resourceName, "backup_strategy.0.keep_days", "10"),
					resource.TestCheckResourceAttr(resourceName, "read_replicas", "4"),
					resource.TestCheckResourceAttr(resourceName, "audit_log_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "sql_filter_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_id",
						"huaweicloud_gaussdb_mysql_parameter_template.test.1", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_name",
						"huaweicloud_gaussdb_mysql_parameter_template.test.1", "name"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.name", "binlog_gtid_simple_recovery"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.value", "ON"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo_update", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value_update"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"table_name_case_sensitivity",
					"enterprise_project_id",
					"password",
					"parameters",
				},
			},
		},
	})
}

func TestAccGaussDBInstance_prePaid(t *testing.T) {
	var (
		instance instances.TaurusDBInstance

		resourceName = "huaweicloud_gaussdb_mysql_instance.test"
		password     = acceptance.RandomPassword()
		rName        = acceptance.RandomAccResourceName()
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acceptance.TestAccPreCheck(t)
			acceptance.TestAccPreCheckChargingMode(t)
		},
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      testAccCheckGaussDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGaussDBInstanceConfig_prePaid(rName, password, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGaussDBInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "auto_renew", "false"),
				),
			},
			{
				Config: testAccGaussDBInstanceConfig_prePaid(rName, password, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGaussDBInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "auto_renew", "true"),
				),
			},
		},
	})
}

func TestAccGaussDBInstance_updateWithEpsId(t *testing.T) {
	var instance instances.TaurusDBInstance

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "huaweicloud_gaussdb_mysql_instance.test"
	srcEPS := acceptance.HW_ENTERPRISE_PROJECT_ID_TEST
	destEPS := acceptance.HW_ENTERPRISE_MIGRATE_PROJECT_ID_TEST

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acceptance.TestAccPreCheck(t)
			acceptance.TestAccPreCheckMigrateEpsID(t)
		},
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      testAccCheckGaussDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGaussDBInstanceConfig_withEpsId(rName, srcEPS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGaussDBInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "enterprise_project_id", srcEPS),
				),
			},
			{
				Config: testAccGaussDBInstanceConfig_withEpsId(rName, destEPS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGaussDBInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "enterprise_project_id", destEPS),
				),
			},
		},
	})
}

func testAccCheckGaussDBInstanceDestroy(s *terraform.State) error {
	cfg := acceptance.TestAccProvider.Meta().(*config.Config)
	client, err := cfg.GaussdbV3Client(acceptance.HW_REGION_NAME)
	if err != nil {
		return fmt.Errorf("error creating GaussDB client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "huaweicloud_gaussdb_mysql_instance" {
			continue
		}

		v, err := instances.Get(client, rs.Primary.ID).Extract()
		if err == nil && v.Id == rs.Primary.ID {
			return fmt.Errorf("instance <%s> still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckGaussDBInstanceExists(n string, instance *instances.TaurusDBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := acceptance.TestAccProvider.Meta().(*config.Config)
		client, err := config.GaussdbV3Client(acceptance.HW_REGION_NAME)
		if err != nil {
			return fmt.Errorf("error creating GaussDB client: %s", err)
		}

		found, err := instances.Get(client, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}
		if found.Id != rs.Primary.ID {
			return fmt.Errorf("instance <%s> not found", rs.Primary.ID)
		}
		instance = found

		return nil
	}
}

func testAccGaussDBInstanceConfig_basic(rName string) string {
	return fmt.Sprintf(`
%[1]s

data "huaweicloud_availability_zones" "test" {}

data "huaweicloud_gaussdb_mysql_flavors" "test" {
  engine  = "gaussdb-mysql"
  version = "8.0"
}

resource "huaweicloud_gaussdb_mysql_parameter_template" "test" {
  count = 2

  name              = "%[2]s_${count.index}"
  datastore_engine  = "gaussdb-mysql"
  datastore_version = "8.0"

  parameter_values = {
    default_authentication_plugin = "sha256_password"
    binlog_gtid_simple_recovery   = "OFF"
  }
}

resource "huaweicloud_gaussdb_mysql_instance" "test" {
  name                     = "%[2]s"
  password                 = "Test@12345678"
  flavor                   = data.huaweicloud_gaussdb_mysql_flavors.test.flavors[0].name
  vpc_id                   = huaweicloud_vpc.test.id
  subnet_id                = huaweicloud_vpc_subnet.test.id
  security_group_id        = huaweicloud_networking_secgroup.test.id
  availability_zone_mode   = "multi"
  master_availability_zone = data.huaweicloud_availability_zones.test.names[0]
  read_replicas            = 2
  enterprise_project_id    = "0"
  sql_filter_enabled       = true
  configuration_id         = huaweicloud_gaussdb_mysql_parameter_template.test[0].id

  parameters {
    name  = "default_authentication_plugin"
    value = "mysql_native_password"
  }

  backup_strategy {
    start_time = "09:00-10:00"
    keep_days  = "7"
  }

  tags = {
    foo = "bar"
    key = "value"
  }
}
`, common.TestBaseNetwork(rName), rName)
}

func testAccGaussDBInstanceConfig_basicUpdate(rName string) string {
	return fmt.Sprintf(`
%[1]s

data "huaweicloud_availability_zones" "test" {}

data "huaweicloud_gaussdb_mysql_flavors" "test" {
  engine  = "gaussdb-mysql"
  version = "8.0"
}

resource "huaweicloud_gaussdb_mysql_parameter_template" "test" {
  count = 2

  name              = "%[2]s_${count.index}"
  datastore_engine  = "gaussdb-mysql"
  datastore_version = "8.0"

  parameter_values = {
    default_authentication_plugin = "sha256_password"
    binlog_gtid_simple_recovery   = "OFF"
  }
}

resource "huaweicloud_gaussdb_mysql_instance" "test" {
  name                     = "%[2]s"
  password                 = "Test@123456789"
  flavor                   = data.huaweicloud_gaussdb_mysql_flavors.test.flavors[1].name
  vpc_id                   = huaweicloud_vpc.test.id
  subnet_id                = huaweicloud_vpc_subnet.test.id
  security_group_id        = huaweicloud_networking_secgroup.test.id
  availability_zone_mode   = "multi"
  master_availability_zone = data.huaweicloud_availability_zones.test.names[0]
  read_replicas            = 4
  enterprise_project_id    = "0"
  audit_log_enabled        = true
  sql_filter_enabled       = false
  configuration_id         = huaweicloud_gaussdb_mysql_parameter_template.test[1].id

  parameters {
    name  = "binlog_gtid_simple_recovery"
    value = "ON"
  }

  backup_strategy {
    start_time = "12:00-13:00"
    keep_days  = "10"
  }

  tags = {
    foo_update = "bar"
    key        = "value_update"
  }
}
`, common.TestBaseNetwork(rName), rName)
}

func testAccGaussDBInstanceConfig_prePaid(rName, password string, isAutoRenew bool) string {
	return fmt.Sprintf(`
%s

data "huaweicloud_availability_zones" "test" {}

resource "huaweicloud_gaussdb_mysql_instance" "test" {
  vpc_id                = huaweicloud_vpc.test.id
  subnet_id             = huaweicloud_vpc_subnet.test.id
  security_group_id     = huaweicloud_networking_secgroup.test.id

  flavor   = "gaussdb.mysql.4xlarge.x86.4"
  name     = "%s"
  password = "%s"

  enterprise_project_id = "0"

  charging_mode = "prePaid"
  period_unit   = "month"
  period        = 1
  auto_renew    = "%v"
}
`, common.TestBaseNetwork(rName), rName, password, isAutoRenew)
}

func testAccGaussDBInstanceConfig_withEpsId(rName, epsId string) string {
	return fmt.Sprintf(`
%s

data "huaweicloud_availability_zones" "test" {}

resource "huaweicloud_gaussdb_mysql_instance" "test" {
  name                  = "%s"
  password              = "Test@12345678"
  flavor                = "gaussdb.mysql.4xlarge.x86.4"
  vpc_id                = huaweicloud_vpc.test.id
  subnet_id             = huaweicloud_vpc_subnet.test.id
  security_group_id     = huaweicloud_networking_secgroup.test.id
  enterprise_project_id = "%s"
  sql_filter_enabled    = true

  tags = {
    foo = "bar"
    key = "value"
  }
}
`, common.TestBaseNetwork(rName), rName, epsId)
}
