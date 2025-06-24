package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSqlRow(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccSqlRowConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.sql_source_row.test",
						tfjsonpath.New("values"),
						knownvalue.MapExact(map[string]knownvalue.Check{
							"id":          knownvalue.StringExact("example-id"),
							"description": knownvalue.StringExact("example description"),
						}),
					),
				},
			},
		},
	})
}

const testAccSqlRowConfig = `
data "sql_source_row" "test" {
  values = {
    "id" = "example-id"
	"description" = "example description"
	}
}
`
