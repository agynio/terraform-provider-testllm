package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-testllm/internal/client"
)

type testResource struct {
	client *client.Client
}

var (
	_ resource.Resource                   = &testResource{}
	_ resource.ResourceWithImportState    = &testResource{}
	_ resource.ResourceWithValidateConfig = &testResource{}
)

type testResourceModel struct {
	ID          types.String    `tfsdk:"id"`
	OrgID       types.String    `tfsdk:"org_id"`
	SuiteID     types.String    `tfsdk:"suite_id"`
	Name        types.String    `tfsdk:"name"`
	Description types.String    `tfsdk:"description"`
	Items       []testItemModel `tfsdk:"items"`
	CreatedAt   types.String    `tfsdk:"created_at"`
	UpdatedAt   types.String    `tfsdk:"updated_at"`
}

type testItemModel struct {
	Type      types.String `tfsdk:"type"`
	Role      types.String `tfsdk:"role"`
	Content   types.String `tfsdk:"content"`
	CallID    types.String `tfsdk:"call_id"`
	FuncName  types.String `tfsdk:"func_name"`
	Arguments types.String `tfsdk:"arguments"`
	Output    types.String `tfsdk:"output"`
}

type itemFieldRule struct {
	Name     string
	Getter   func(testItemModel) types.String
	Required bool
}

var testItemValidationRules = map[string][]itemFieldRule{
	"message": {
		{Name: "role", Getter: func(item testItemModel) types.String { return item.Role }, Required: true},
		{Name: "content", Getter: func(item testItemModel) types.String { return item.Content }, Required: true},
		{Name: "call_id", Getter: func(item testItemModel) types.String { return item.CallID }},
		{Name: "func_name", Getter: func(item testItemModel) types.String { return item.FuncName }},
		{Name: "arguments", Getter: func(item testItemModel) types.String { return item.Arguments }},
		{Name: "output", Getter: func(item testItemModel) types.String { return item.Output }},
	},
	"function_call": {
		{Name: "call_id", Getter: func(item testItemModel) types.String { return item.CallID }, Required: true},
		{Name: "func_name", Getter: func(item testItemModel) types.String { return item.FuncName }, Required: true},
		{Name: "arguments", Getter: func(item testItemModel) types.String { return item.Arguments }, Required: true},
		{Name: "role", Getter: func(item testItemModel) types.String { return item.Role }},
		{Name: "content", Getter: func(item testItemModel) types.String { return item.Content }},
		{Name: "output", Getter: func(item testItemModel) types.String { return item.Output }},
	},
	"function_call_output": {
		{Name: "call_id", Getter: func(item testItemModel) types.String { return item.CallID }, Required: true},
		{Name: "output", Getter: func(item testItemModel) types.String { return item.Output }, Required: true},
		{Name: "role", Getter: func(item testItemModel) types.String { return item.Role }},
		{Name: "content", Getter: func(item testItemModel) types.String { return item.Content }},
		{Name: "func_name", Getter: func(item testItemModel) types.String { return item.FuncName }},
		{Name: "arguments", Getter: func(item testItemModel) types.String { return item.Arguments }},
	},
}

func NewTestResource() resource.Resource {
	return &testResource{}
}

func (r *testResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test"
}

func (r *testResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"suite_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"items": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("message", "function_call", "function_call_output"),
							},
						},
						"role": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.OneOf("user", "system", "developer", "assistant"),
							},
						},
						"content": schema.StringAttribute{
							Optional: true,
						},
						"call_id": schema.StringAttribute{
							Optional: true,
						},
						"func_name": schema.StringAttribute{
							Optional: true,
						},
						"arguments": schema.StringAttribute{
							Optional: true,
						},
						"output": schema.StringAttribute{
							Optional: true,
						},
					},
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

func (r *testResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *testResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config testResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for index, item := range config.Items {
		if item.Type.IsUnknown() || item.Type.IsNull() {
			continue
		}
		itemType := item.Type.ValueString()
		itemPath := path.Root("items").AtListIndex(index)
		rules, ok := testItemValidationRules[itemType]
		if !ok {
			resp.Diagnostics.AddAttributeError(itemPath.AtName("type"), "Invalid item type", fmt.Sprintf("Unsupported item type %q.", itemType))
			continue
		}
		for _, rule := range rules {
			attrPath := itemPath.AtName(rule.Name)
			value := rule.Getter(item)
			if rule.Required {
				validateRequiredString(&resp.Diagnostics, value, attrPath, itemType)
				continue
			}
			validateUnexpectedString(&resp.Diagnostics, value, attrPath, itemType)
		}
	}
}

func (r *testResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan testResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, diags := expandTestItems(plan.Items)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateTest(ctx, plan.OrgID.ValueString(), plan.SuiteID.ValueString(), plan.Name.ValueString(), plan.Description.ValueString(), items)
	if err != nil {
		resp.Diagnostics.AddError("Error creating test", err.Error())
		return
	}

	itemModels, diags := flattenTestItems(created.Items)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := testResourceModel{
		ID:          types.StringValue(created.ID),
		OrgID:       plan.OrgID,
		SuiteID:     types.StringValue(created.TestSuiteID),
		Name:        types.StringValue(created.Name),
		Description: types.StringValue(created.Description),
		Items:       itemModels,
		CreatedAt:   types.StringValue(created.CreatedAt),
		UpdatedAt:   types.StringValue(created.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *testResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state testResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	read, err := r.client.GetTest(ctx, state.OrgID.ValueString(), state.SuiteID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading test", err.Error())
		return
	}

	itemModels, diags := flattenTestItems(read.Items)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state = testResourceModel{
		ID:          types.StringValue(read.ID),
		OrgID:       state.OrgID,
		SuiteID:     types.StringValue(read.TestSuiteID),
		Name:        types.StringValue(read.Name),
		Description: types.StringValue(read.Description),
		Items:       itemModels,
		CreatedAt:   types.StringValue(read.CreatedAt),
		UpdatedAt:   types.StringValue(read.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *testResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan testResourceModel
	var state testResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, diags := expandTestItems(plan.Items)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateTest(ctx, state.OrgID.ValueString(), state.SuiteID.ValueString(), state.ID.ValueString(), plan.Name.ValueString(), plan.Description.ValueString(), items)
	if err != nil {
		resp.Diagnostics.AddError("Error updating test", err.Error())
		return
	}

	itemModels, diags := flattenTestItems(updated.Items)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state = testResourceModel{
		ID:          types.StringValue(updated.ID),
		OrgID:       state.OrgID,
		SuiteID:     types.StringValue(updated.TestSuiteID),
		Name:        types.StringValue(updated.Name),
		Description: types.StringValue(updated.Description),
		Items:       itemModels,
		CreatedAt:   types.StringValue(updated.CreatedAt),
		UpdatedAt:   types.StringValue(updated.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *testResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state testResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTest(ctx, state.OrgID.ValueString(), state.SuiteID.ValueString(), state.ID.ValueString()); err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting test", err.Error())
		return
	}
}

func (r *testResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import identifier with format: <org_id>/<suite_id>/<test_id>. Got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("suite_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[2])...)
}

func validateRequiredString(diags *diag.Diagnostics, value types.String, attrPath path.Path, itemType string) {
	if value.IsUnknown() {
		return
	}
	if value.IsNull() {
		diags.AddAttributeError(attrPath, "Missing required item attribute", fmt.Sprintf("%s is required when type is %q.", attrPath.String(), itemType))
	}
}

func validateUnexpectedString(diags *diag.Diagnostics, value types.String, attrPath path.Path, itemType string) {
	if value.IsUnknown() || value.IsNull() {
		return
	}
	diags.AddAttributeError(attrPath, "Unexpected item attribute", fmt.Sprintf("%s must not be set when type is %q.", attrPath.String(), itemType))
}

func expandTestItems(items []testItemModel) ([]client.TestItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(items) == 0 {
		return nil, diags
	}

	expanded := make([]client.TestItem, 0, len(items))
	for index, item := range items {
		if item.Type.IsUnknown() || item.Type.IsNull() {
			diags.AddAttributeError(path.Root("items").AtListIndex(index).AtName("type"), "Invalid item type", "Each item must include a known type.")
			return nil, diags
		}

		itemType := item.Type.ValueString()
		switch itemType {
		case "message":
			messageItem, err := client.NewMessageItem(item.Role.ValueString(), item.Content.ValueString())
			if err != nil {
				diags.AddError("Error building message item", err.Error())
				return nil, diags
			}
			expanded = append(expanded, messageItem)
		case "function_call":
			callItem, err := client.NewFunctionCallItem(item.CallID.ValueString(), item.FuncName.ValueString(), item.Arguments.ValueString())
			if err != nil {
				diags.AddError("Error building function_call item", err.Error())
				return nil, diags
			}
			expanded = append(expanded, callItem)
		case "function_call_output":
			outputItem, err := client.NewFunctionCallOutputItem(item.CallID.ValueString(), item.Output.ValueString())
			if err != nil {
				diags.AddError("Error building function_call_output item", err.Error())
				return nil, diags
			}
			expanded = append(expanded, outputItem)
		default:
			diags.AddAttributeError(path.Root("items").AtListIndex(index).AtName("type"), "Invalid item type", fmt.Sprintf("Unsupported item type %q.", itemType))
			return nil, diags
		}
	}

	return expanded, diags
}

func flattenTestItems(items []client.TestItem) ([]testItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(items) == 0 {
		return nil, diags
	}

	flattened := make([]testItemModel, 0, len(items))
	for index, item := range items {
		switch item.Type {
		case "message":
			role, content, err := client.ParseMessageContent(item)
			if err != nil {
				diags.AddError("Error parsing message item", err.Error())
				return nil, diags
			}
			flattened = append(flattened, testItemModel{
				Type:      types.StringValue("message"),
				Role:      types.StringValue(role),
				Content:   types.StringValue(content),
				CallID:    types.StringNull(),
				FuncName:  types.StringNull(),
				Arguments: types.StringNull(),
				Output:    types.StringNull(),
			})
		case "function_call":
			callID, name, arguments, err := client.ParseFunctionCallContent(item)
			if err != nil {
				diags.AddError("Error parsing function_call item", err.Error())
				return nil, diags
			}
			flattened = append(flattened, testItemModel{
				Type:      types.StringValue("function_call"),
				Role:      types.StringNull(),
				Content:   types.StringNull(),
				CallID:    types.StringValue(callID),
				FuncName:  types.StringValue(name),
				Arguments: types.StringValue(arguments),
				Output:    types.StringNull(),
			})
		case "function_call_output":
			callID, output, err := client.ParseFunctionCallOutputContent(item)
			if err != nil {
				diags.AddError("Error parsing function_call_output item", err.Error())
				return nil, diags
			}
			flattened = append(flattened, testItemModel{
				Type:      types.StringValue("function_call_output"),
				Role:      types.StringNull(),
				Content:   types.StringNull(),
				CallID:    types.StringValue(callID),
				FuncName:  types.StringNull(),
				Arguments: types.StringNull(),
				Output:    types.StringValue(output),
			})
		default:
			diags.AddAttributeError(path.Root("items").AtListIndex(index).AtName("type"), "Invalid item type", fmt.Sprintf("Unsupported item type %q.", item.Type))
			return nil, diags
		}
	}

	return flattened, diags
}
