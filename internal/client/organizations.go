package client

import (
	"context"
	"fmt"
	"net/http"
)

// Organization represents an organization in TestLLM.
type Organization struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type organizationCreateRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type organizationUpdateRequest struct {
	Name string `json:"name"`
}

func (c *Client) CreateOrganization(ctx context.Context, name, slug string) (*Organization, error) {
	request := organizationCreateRequest{
		Name: name,
		Slug: slug,
	}
	var response Organization
	if err := c.doRequest(ctx, http.MethodPost, "/api/orgs", request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) GetOrganization(ctx context.Context, orgID string) (*Organization, error) {
	var response Organization
	path := fmt.Sprintf("/api/orgs/%s", orgID)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) UpdateOrganization(ctx context.Context, orgID, name string) (*Organization, error) {
	request := organizationUpdateRequest{
		Name: name,
	}
	var response Organization
	path := fmt.Sprintf("/api/orgs/%s", orgID)
	if err := c.doRequest(ctx, http.MethodPatch, path, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) DeleteOrganization(ctx context.Context, orgID string) error {
	path := fmt.Sprintf("/api/orgs/%s", orgID)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
