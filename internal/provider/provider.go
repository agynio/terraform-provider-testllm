package provider

import (
	"context"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-testllm/internal/client"
)

type testllmProvider struct {
	version string
}

type providerConfig struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

// New returns a new provider instance.
func New() provider.Provider {
	return &testllmProvider{version: "dev"}
}

func (p *testllmProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "testllm"
	resp.Version = p.version
}

func (p *testllmProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "TestLLM API base URL. Defaults to https://testllm.dev. May also be set via the TESTLLM_HOST environment variable.",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "API authentication token. May also be set via the TESTLLM_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *testllmProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown TestLLM host",
			"The provider cannot create the TestLLM client because the host value is unknown. Set the host in configuration or via TESTLLM_HOST.",
		)
		return
	}
	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown TestLLM token",
			"The provider cannot create the TestLLM client because the token value is unknown. Set the token in configuration or via TESTLLM_TOKEN.",
		)
		return
	}

	host := config.Host.ValueString()
	if host == "" {
		host = os.Getenv("TESTLLM_HOST")
	}
	if host == "" {
		host = "https://testllm.dev"
	}

	token := config.Token.ValueString()
	if token == "" {
		token = os.Getenv("TESTLLM_TOKEN")
	}
	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing TestLLM token",
			"The provider cannot create the TestLLM client because the token value is missing. Set the token in configuration or via TESTLLM_TOKEN.",
		)
		return
	}

	parsedURL, err := url.Parse(host)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Invalid TestLLM host",
			"The provider cannot create the TestLLM client because the host value is invalid.",
		)
		return
	}

	client := client.New(parsedURL, token)
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *testllmProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrganizationResource,
		NewTestSuiteResource,
		NewTestResource,
	}
}

func (p *testllmProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrganizationDataSource,
	}
}
