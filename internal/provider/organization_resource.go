package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-testllm/internal/client"
)

type organizationResource struct {
	client *client.Client
}

var (
	_ resource.Resource                = &organizationResource{}
	_ resource.ResourceWithImportState = &organizationResource{}
)

type organizationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Slug      types.String `tfsdk:"slug"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func NewOrganizationResource() resource.Resource {
	return &organizationResource{}
}

func (r *organizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *organizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"slug": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *organizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *client.Client")
		return
	}
	r.client = client
}

func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.CreateOrganization(ctx, plan.Name.ValueString(), plan.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating organization", err.Error())
		return
	}

	state := organizationResourceModel{
		ID:        types.StringValue(org.ID),
		Name:      types.StringValue(org.Name),
		Slug:      types.StringValue(org.Slug),
		CreatedAt: types.StringValue(org.CreatedAt),
		UpdatedAt: types.StringValue(org.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.GetOrganization(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading organization", err.Error())
		return
	}

	state = organizationResourceModel{
		ID:        types.StringValue(org.ID),
		Name:      types.StringValue(org.Name),
		Slug:      types.StringValue(org.Slug),
		CreatedAt: types.StringValue(org.CreatedAt),
		UpdatedAt: types.StringValue(org.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationResourceModel
	var state organizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.UpdateOrganization(ctx, state.ID.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating organization", err.Error())
		return
	}

	state = organizationResourceModel{
		ID:        types.StringValue(org.ID),
		Name:      types.StringValue(org.Name),
		Slug:      types.StringValue(org.Slug),
		CreatedAt: types.StringValue(org.CreatedAt),
		UpdatedAt: types.StringValue(org.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteOrganization(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting organization", err.Error())
		return
	}
}

func (r *organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
