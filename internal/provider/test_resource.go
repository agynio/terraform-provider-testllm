package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	Type          types.String `tfsdk:"type"`
	Role          types.String `tfsdk:"role"`
	Content       types.String `tfsdk:"content"`
	Text          types.String `tfsdk:"text"`
	ContentBlocks types.String `tfsdk:"content_blocks"`
	AnyRole       types.Bool   `tfsdk:"any_role"`
	AnyContent    types.Bool   `tfsdk:"any_content"`
	Repeat        types.Bool   `tfsdk:"repeat"`
	CallID        types.String `tfsdk:"call_id"`
	FuncName      types.String `tfsdk:"func_name"`
	Arguments     types.String `tfsdk:"arguments"`
	Output        types.String `tfsdk:"output"`
}

type itemFieldRule struct {
	Name     string
	Getter   func(testItemModel) types.String
	Required bool
}

type itemBoolFieldRule struct {
	Name   string
	Getter func(testItemModel) types.Bool
}

var testItemValidationRules = map[string][]itemFieldRule{
	"message": {
		{Name: "role", Getter: func(item testItemModel) types.String { return item.Role }, Required: true},
		{Name: "content", Getter: func(item testItemModel) types.String { return item.Content }, Required: true},
		{Name: "call_id", Getter: func(item testItemModel) types.String { return item.CallID }},
		{Name: "func_name", Getter: func(item testItemModel) types.String { return item.FuncName }},
		{Name: "arguments", Getter: func(item testItemModel) types.String { return item.Arguments }},
		{Name: "output", Getter: func(item testItemModel) types.String { return item.Output }},
		{Name: "text", Getter: func(item testItemModel) types.String { return item.Text }},
		{Name: "content_blocks", Getter: func(item testItemModel) types.String { return item.ContentBlocks }},
	},
	"function_call": {
		{Name: "call_id", Getter: func(item testItemModel) types.String { return item.CallID }, Required: true},
		{Name: "func_name", Getter: func(item testItemModel) types.String { return item.FuncName }, Required: true},
		{Name: "arguments", Getter: func(item testItemModel) types.String { return item.Arguments }, Required: true},
		{Name: "role", Getter: func(item testItemModel) types.String { return item.Role }},
		{Name: "content", Getter: func(item testItemModel) types.String { return item.Content }},
		{Name: "output", Getter: func(item testItemModel) types.String { return item.Output }},
		{Name: "text", Getter: func(item testItemModel) types.String { return item.Text }},
		{Name: "content_blocks", Getter: func(item testItemModel) types.String { return item.ContentBlocks }},
	},
	"function_call_output": {
		{Name: "call_id", Getter: func(item testItemModel) types.String { return item.CallID }, Required: true},
		{Name: "output", Getter: func(item testItemModel) types.String { return item.Output }, Required: true},
		{Name: "role", Getter: func(item testItemModel) types.String { return item.Role }},
		{Name: "content", Getter: func(item testItemModel) types.String { return item.Content }},
		{Name: "func_name", Getter: func(item testItemModel) types.String { return item.FuncName }},
		{Name: "arguments", Getter: func(item testItemModel) types.String { return item.Arguments }},
		{Name: "text", Getter: func(item testItemModel) types.String { return item.Text }},
		{Name: "content_blocks", Getter: func(item testItemModel) types.String { return item.ContentBlocks }},
	},
	"anthropic_system": {
		{Name: "role", Getter: func(item testItemModel) types.String { return item.Role }},
		{Name: "content", Getter: func(item testItemModel) types.String { return item.Content }},
		{Name: "call_id", Getter: func(item testItemModel) types.String { return item.CallID }},
		{Name: "func_name", Getter: func(item testItemModel) types.String { return item.FuncName }},
		{Name: "arguments", Getter: func(item testItemModel) types.String { return item.Arguments }},
		{Name: "output", Getter: func(item testItemModel) types.String { return item.Output }},
	},
	"anthropic_message": {
		{Name: "role", Getter: func(item testItemModel) types.String { return item.Role }, Required: true},
		{Name: "text", Getter: func(item testItemModel) types.String { return item.Text }},
		{Name: "call_id", Getter: func(item testItemModel) types.String { return item.CallID }},
		{Name: "func_name", Getter: func(item testItemModel) types.String { return item.FuncName }},
		{Name: "arguments", Getter: func(item testItemModel) types.String { return item.Arguments }},
		{Name: "output", Getter: func(item testItemModel) types.String { return item.Output }},
	},
}

var testItemBoolValidationRules = []itemBoolFieldRule{
	{Name: "any_role", Getter: func(item testItemModel) types.Bool { return item.AnyRole }},
	{Name: "any_content", Getter: func(item testItemModel) types.Bool { return item.AnyContent }},
	{Name: "repeat", Getter: func(item testItemModel) types.Bool { return item.Repeat }},
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
				Description: "Unique identifier for the test.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Description: "Organization ID that owns the test.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"suite_id": schema.StringAttribute{
				Description: "Test suite ID that owns the test.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Display name for the test.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the test.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"items": schema.ListNestedAttribute{
				Description: "Ordered list of test items that define the test flow.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "Item type. One of message, function_call, function_call_output, anthropic_system, or anthropic_message.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("message", "function_call", "function_call_output", "anthropic_system", "anthropic_message"),
							},
						},
						"role": schema.StringAttribute{
							Description: "Role for message and anthropic_message items (user, system, developer, assistant).",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("user", "system", "developer", "assistant"),
							},
						},
						"content": schema.StringAttribute{
							Description: "Content for message items.",
							Optional:    true,
						},
						"text": schema.StringAttribute{
							Description: "Text content for anthropic_system items.",
							Optional:    true,
						},
						"content_blocks": schema.StringAttribute{
							Description: "JSON-encoded array of Anthropic content blocks.",
							Optional:    true,
						},
						"any_role": schema.BoolAttribute{
							Description: "Whether any role is accepted for message items.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"any_content": schema.BoolAttribute{
							Description: "Whether any content is accepted for message items.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"repeat": schema.BoolAttribute{
							Description: "Whether the message item can repeat.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"call_id": schema.StringAttribute{
							Description: "Function call identifier for function_call and function_call_output items.",
							Optional:    true,
						},
						"func_name": schema.StringAttribute{
							Description: "Function name for function_call items.",
							Optional:    true,
						},
						"arguments": schema.StringAttribute{
							Description: "JSON-encoded arguments for function_call items.",
							Optional:    true,
						},
						"output": schema.StringAttribute{
							Description: "Output content for function_call_output items.",
							Optional:    true,
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the test was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the test was last updated.",
				Computed:    true,
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

		for _, rule := range testItemBoolValidationRules {
			attrPath := itemPath.AtName(rule.Name)
			value := rule.Getter(item)
			if value.IsUnknown() || value.IsNull() || !value.ValueBool() {
				continue
			}
			allowed := itemType == "message"
			if rule.Name == "any_content" {
				allowed = itemType == "message" || itemType == "anthropic_message" || itemType == "anthropic_system"
			}
			if !allowed {
				resp.Diagnostics.AddAttributeError(attrPath, "Unexpected item attribute", fmt.Sprintf("%s must not be true when type is %q.", attrPath.String(), itemType))
				continue
			}
			if item.Role.IsUnknown() || item.Role.IsNull() {
				continue
			}
			if item.Role.ValueString() == "assistant" && itemType == "message" {
				resp.Diagnostics.AddAttributeError(attrPath, "Unexpected item attribute", fmt.Sprintf("%s must not be true when type is %q and role is %q.", attrPath.String(), itemType, item.Role.ValueString()))
			}
		}

		switch itemType {
		case "anthropic_system":
			if !item.Text.IsUnknown() && !item.ContentBlocks.IsUnknown() {
				textSet := !item.Text.IsNull()
				blocksSet := !item.ContentBlocks.IsNull()
				if textSet == blocksSet {
					resp.Diagnostics.AddAttributeError(
						itemPath.AtName("text"),
						"Invalid item attributes",
						fmt.Sprintf("Exactly one of %s or %s must be set when type is %q.", itemPath.AtName("text").String(), itemPath.AtName("content_blocks").String(), itemType),
					)
				}
			}
			validateContentBlocks(&resp.Diagnostics, item.ContentBlocks, itemPath.AtName("content_blocks"))
		case "anthropic_message":
			if !item.Content.IsUnknown() && !item.ContentBlocks.IsUnknown() {
				contentSet := !item.Content.IsNull()
				blocksSet := !item.ContentBlocks.IsNull()
				if contentSet == blocksSet {
					resp.Diagnostics.AddAttributeError(
						itemPath.AtName("content"),
						"Invalid item attributes",
						fmt.Sprintf("Exactly one of %s or %s must be set when type is %q.", itemPath.AtName("content").String(), itemPath.AtName("content_blocks").String(), itemType),
					)
				}
			}
			if !item.Role.IsUnknown() && !item.Role.IsNull() {
				role := item.Role.ValueString()
				if role != "user" && role != "assistant" {
					resp.Diagnostics.AddAttributeError(
						itemPath.AtName("role"),
						"Invalid item role",
						fmt.Sprintf("%s must be either \"user\" or \"assistant\" when type is %q.", itemPath.AtName("role").String(), itemType),
					)
				}
			}
			validateContentBlocks(&resp.Diagnostics, item.ContentBlocks, itemPath.AtName("content_blocks"))
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

func validateContentBlocks(diags *diag.Diagnostics, value types.String, attrPath path.Path) {
	if value.IsUnknown() || value.IsNull() {
		return
	}

	var blocks []struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(value.ValueString()), &blocks); err != nil {
		diags.AddAttributeError(attrPath, "Invalid content_blocks", fmt.Sprintf("%s must be a JSON-encoded array of objects with a type field of \"text\", \"tool_use\", or \"tool_result\": %s.", attrPath.String(), err.Error()))
		return
	}

	allowedTypes := map[string]struct{}{
		"text":        {},
		"tool_use":    {},
		"tool_result": {},
	}
	for _, block := range blocks {
		if _, ok := allowedTypes[block.Type]; !ok {
			diags.AddAttributeError(attrPath, "Invalid content_blocks", fmt.Sprintf("%s must be a JSON-encoded array of objects with a type field of \"text\", \"tool_use\", or \"tool_result\".", attrPath.String()))
			return
		}
	}
}

func boolPointerFromValue(value types.Bool) *bool {
	if value.IsUnknown() || value.IsNull() || !value.ValueBool() {
		return nil
	}
	boolValue := true
	return &boolValue
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
			anyRole := boolPointerFromValue(item.AnyRole)
			anyContent := boolPointerFromValue(item.AnyContent)
			repeat := boolPointerFromValue(item.Repeat)
			messageItem, err := client.NewMessageItem(item.Role.ValueString(), item.Content.ValueString(), anyRole, anyContent, repeat)
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
		case "anthropic_system":
			textSet := !item.Text.IsNull() && !item.Text.IsUnknown()
			blocksSet := !item.ContentBlocks.IsNull() && !item.ContentBlocks.IsUnknown()
			if textSet == blocksSet {
				diags.AddAttributeError(
					path.Root("items").AtListIndex(index).AtName("text"),
					"Invalid item attributes",
					fmt.Sprintf("Exactly one of %s or %s must be set when type is %q.", path.Root("items").AtListIndex(index).AtName("text").String(), path.Root("items").AtListIndex(index).AtName("content_blocks").String(), itemType),
				)
				return nil, diags
			}
			var systemItem client.TestItem
			var err error
			anyContent := boolPointerFromValue(item.AnyContent)
			if textSet {
				systemItem, err = client.NewAnthropicSystemTextItem(item.Text.ValueString(), anyContent)
			} else {
				systemItem, err = client.NewAnthropicSystemBlocksItem(json.RawMessage(item.ContentBlocks.ValueString()), anyContent)
			}
			if err != nil {
				diags.AddError("Error building anthropic_system item", err.Error())
				return nil, diags
			}
			expanded = append(expanded, systemItem)
		case "anthropic_message":
			contentSet := !item.Content.IsNull() && !item.Content.IsUnknown()
			blocksSet := !item.ContentBlocks.IsNull() && !item.ContentBlocks.IsUnknown()
			if contentSet == blocksSet {
				diags.AddAttributeError(
					path.Root("items").AtListIndex(index).AtName("content"),
					"Invalid item attributes",
					fmt.Sprintf("Exactly one of %s or %s must be set when type is %q.", path.Root("items").AtListIndex(index).AtName("content").String(), path.Root("items").AtListIndex(index).AtName("content_blocks").String(), itemType),
				)
				return nil, diags
			}
			if item.Role.IsUnknown() || item.Role.IsNull() {
				diags.AddAttributeError(path.Root("items").AtListIndex(index).AtName("role"), "Missing required item attribute", fmt.Sprintf("%s is required when type is %q.", path.Root("items").AtListIndex(index).AtName("role").String(), itemType))
				return nil, diags
			}
			var messageItem client.TestItem
			var err error
			anyContent := boolPointerFromValue(item.AnyContent)
			if contentSet {
				messageItem, err = client.NewAnthropicMessageStringItem(item.Role.ValueString(), item.Content.ValueString(), anyContent)
			} else {
				messageItem, err = client.NewAnthropicMessageBlocksItem(item.Role.ValueString(), json.RawMessage(item.ContentBlocks.ValueString()), anyContent)
			}
			if err != nil {
				diags.AddError("Error building anthropic_message item", err.Error())
				return nil, diags
			}
			expanded = append(expanded, messageItem)
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
			messageContent, err := client.ParseMessageContent(item)
			if err != nil {
				diags.AddError("Error parsing message item", err.Error())
				return nil, diags
			}
			flattened = append(flattened, testItemModel{
				Type:          types.StringValue("message"),
				Role:          types.StringValue(messageContent.Role),
				Content:       types.StringValue(messageContent.Content),
				Text:          types.StringNull(),
				ContentBlocks: types.StringNull(),
				AnyRole:       types.BoolValue(messageContent.AnyRole),
				AnyContent:    types.BoolValue(messageContent.AnyContent),
				Repeat:        types.BoolValue(messageContent.Repeat),
				CallID:        types.StringNull(),
				FuncName:      types.StringNull(),
				Arguments:     types.StringNull(),
				Output:        types.StringNull(),
			})
		case "function_call":
			callID, name, arguments, err := client.ParseFunctionCallContent(item)
			if err != nil {
				diags.AddError("Error parsing function_call item", err.Error())
				return nil, diags
			}
			flattened = append(flattened, testItemModel{
				Type:          types.StringValue("function_call"),
				Role:          types.StringNull(),
				Content:       types.StringNull(),
				Text:          types.StringNull(),
				ContentBlocks: types.StringNull(),
				AnyRole:       types.BoolValue(false),
				AnyContent:    types.BoolValue(false),
				Repeat:        types.BoolValue(false),
				CallID:        types.StringValue(callID),
				FuncName:      types.StringValue(name),
				Arguments:     types.StringValue(arguments),
				Output:        types.StringNull(),
			})
		case "function_call_output":
			callID, output, err := client.ParseFunctionCallOutputContent(item)
			if err != nil {
				diags.AddError("Error parsing function_call_output item", err.Error())
				return nil, diags
			}
			flattened = append(flattened, testItemModel{
				Type:          types.StringValue("function_call_output"),
				Role:          types.StringNull(),
				Content:       types.StringNull(),
				Text:          types.StringNull(),
				ContentBlocks: types.StringNull(),
				AnyRole:       types.BoolValue(false),
				AnyContent:    types.BoolValue(false),
				Repeat:        types.BoolValue(false),
				CallID:        types.StringValue(callID),
				FuncName:      types.StringNull(),
				Arguments:     types.StringNull(),
				Output:        types.StringValue(output),
			})
		case "anthropic_system":
			systemContent, err := client.ParseAnthropicSystemContent(item)
			if err != nil {
				diags.AddError("Error parsing anthropic_system item", err.Error())
				return nil, diags
			}
			textValue := types.StringNull()
			blocksValue := types.StringNull()
			if systemContent.Blocks != nil {
				blocksValue = types.StringValue(string(systemContent.Blocks))
			} else {
				textValue = types.StringValue(systemContent.Text)
			}
			flattened = append(flattened, testItemModel{
				Type:          types.StringValue("anthropic_system"),
				Role:          types.StringNull(),
				Content:       types.StringNull(),
				Text:          textValue,
				ContentBlocks: blocksValue,
				AnyRole:       types.BoolValue(false),
				AnyContent:    types.BoolValue(systemContent.AnyContent),
				Repeat:        types.BoolValue(false),
				CallID:        types.StringNull(),
				FuncName:      types.StringNull(),
				Arguments:     types.StringNull(),
				Output:        types.StringNull(),
			})
		case "anthropic_message":
			messageContent, err := client.ParseAnthropicMessageContent(item)
			if err != nil {
				diags.AddError("Error parsing anthropic_message item", err.Error())
				return nil, diags
			}
			contentValue := types.StringNull()
			blocksValue := types.StringNull()
			if messageContent.ContentBlocks != nil {
				blocksValue = types.StringValue(string(messageContent.ContentBlocks))
			} else {
				contentValue = types.StringValue(messageContent.Content)
			}
			flattened = append(flattened, testItemModel{
				Type:          types.StringValue("anthropic_message"),
				Role:          types.StringValue(messageContent.Role),
				Content:       contentValue,
				Text:          types.StringNull(),
				ContentBlocks: blocksValue,
				AnyRole:       types.BoolValue(false),
				AnyContent:    types.BoolValue(messageContent.AnyContent),
				Repeat:        types.BoolValue(false),
				CallID:        types.StringNull(),
				FuncName:      types.StringNull(),
				Arguments:     types.StringNull(),
				Output:        types.StringNull(),
			})
		default:
			diags.AddAttributeError(path.Root("items").AtListIndex(index).AtName("type"), "Invalid item type", fmt.Sprintf("Unsupported item type %q.", item.Type))
			return nil, diags
		}
	}

	return flattened, diags
}
