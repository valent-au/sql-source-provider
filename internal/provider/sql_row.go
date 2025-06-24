// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"database/sql"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SqlSourceRowDataSource{}

func NewSqlSourceRowDataSource() datasource.DataSource {
	return &SqlSourceRowDataSource{}
}

// ExampleDataSource defines the data source implementation.
type SqlSourceRowDataSource struct {
	client *sql.Conn
}

// ExampleDataSourceModel describes the data source data model.
type SqlSourceRowDataSourceModel struct {
	QueryString types.String `tfsdk:"query_string"`
	Value       types.Map    `tfsdk:"value"`
}

func (d *SqlSourceRowDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_source_row"
}

func (d *SqlSourceRowDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SQL data source.",

		Attributes: map[string]schema.Attribute{
			"query_string": schema.StringAttribute{
				MarkdownDescription: "The SQL query string to execute. The query must return a single row.",
				Optional:            false,
			},
			"value": schema.MapAttribute{
				MarkdownDescription: "The result of the SQL query.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *SqlSourceRowDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sql.Conn)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *SqlSourceRowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SqlSourceRowDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Make a request to the SQL database using the query string provided in the data source configuration.
	rows, err := d.client.QueryContext(ctx, data.QueryString.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"SQL Query Error",
			fmt.Sprintf("Failed to execute SQL query: %s", err),
		)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		resp.Diagnostics.AddError(
			"SQL Columns Error",
			fmt.Sprintf("Failed to get columns: %s", err),
		)
		return
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	result := make(map[string]attr.Value, len(columns))

	if rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			resp.Diagnostics.AddError(
				"SQL Scan Error",
				fmt.Sprintf("Failed to scan row: %s", err),
			)
			return
		}
		for i, col := range columns {
			var v string
			val := values[i]
			if val != nil {
				v = fmt.Sprintf("%v", val)
			} else {
				v = ""
			}
			result[col] = types.StringValue(v)
		}
	}

	if rows.Next() {
		resp.Diagnostics.AddError(
			"SQL Query Error",
			"Query returned more than one row, expected a single row.",
		)
		return
	}

	data.Value = types.MapValueMust(types.StringType, result)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
