package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-testllm/internal/client"
)

type organizationDataSource struct {
	client *client.Client
}

type organizationDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Slug      types.String `tfsdk:"slug"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

var _ datasource.DataSource = &organizationDataSource{}

func NewOrganizationDataSource() datasource.DataSource {
	return &organizationDataSource{}
}

func (d *organizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *organizationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"slug": schema.StringAttribute{
				Description: "The unique slug of the organization to look up.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Unique identifier for the organization.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Display name for the organization.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the organization was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the organization was last updated.",
				Computed:    true,
			},
		},
	}
}

func (d *organizationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clientInstance, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *client.Client")
		return
	}
	d.client = clientInstance
}

func (d *organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	organizations, err := d.client.ListOrganizations(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing organizations", err.Error())
		return
	}

	var matched *client.Organization
	for index := range organizations {
		if organizations[index].Slug == config.Slug.ValueString() {
			matched = &organizations[index]
			break
		}
	}
	if matched == nil {
		resp.Diagnostics.AddError(
			"Organization not found",
			fmt.Sprintf("No organization found with slug %q.", config.Slug.ValueString()),
		)
		return
	}

	state := organizationDataSourceModel{
		ID:        types.StringValue(matched.ID),
		Name:      types.StringValue(matched.Name),
		Slug:      types.StringValue(matched.Slug),
		CreatedAt: types.StringValue(matched.CreatedAt),
		UpdatedAt: types.StringValue(matched.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
