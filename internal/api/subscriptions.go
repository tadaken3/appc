package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kenta-tanaka/appc/internal/client"
)

// --- JSON:API resource types ---

type SubscriptionGroupResource struct {
	Type       string                      `json:"type"`
	ID         string                      `json:"id"`
	Attributes SubscriptionGroupAttributes `json:"attributes"`
}

type SubscriptionGroupAttributes struct {
	ReferenceName string `json:"referenceName"`
}

type SubscriptionResource struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id"`
	Attributes SubscriptionAttributes `json:"attributes"`
}

type SubscriptionAttributes struct {
	Name                      string `json:"name"`
	ProductID                 string `json:"productId"`
	State                     string `json:"state"`
	SubscriptionPeriod        string `json:"subscriptionPeriod"`
	GroupLevel                int    `json:"groupLevel"`
	ReviewNote                string `json:"reviewNote"`
	FamilySharable            bool   `json:"familySharable"`
	AvailableInAllTerritories bool   `json:"availableInAllTerritories"`
}

// --- Params ---

type SubscriptionParams struct {
	AppID   string
	GroupID string
}

// --- Output model ---

type SubscriptionGroup struct {
	ID            string `json:"id"`
	ReferenceName string `json:"reference_name"`
}

type SubscriptionInfo struct {
	GroupID            string `json:"group_id"`
	GroupName          string `json:"group_name"`
	SubscriptionID     string `json:"subscription_id"`
	Name               string `json:"name"`
	ProductID          string `json:"product_id"`
	State              string `json:"state"`
	SubscriptionPeriod string `json:"subscription_period"`
	GroupLevel         int    `json:"group_level"`
}

func (s SubscriptionInfo) CSVHeaders() []string {
	return []string{"group_id", "group_name", "subscription_id", "name", "product_id", "state", "subscription_period", "group_level"}
}

func (s SubscriptionInfo) CSVRow() []string {
	return []string{s.GroupID, s.GroupName, s.SubscriptionID, s.Name, s.ProductID, s.State, s.SubscriptionPeriod, intToStr(s.GroupLevel)}
}

// --- API functions ---

// ListSubscriptionGroups fetches subscription groups for an app.
func ListSubscriptionGroups(ctx context.Context, c *client.Client, appID string) ([]SubscriptionGroup, error) {
	var groups []SubscriptionGroup
	path := fmt.Sprintf("/v1/apps/%s/subscriptionGroups", appID)

	for path != "" {
		resp, err := c.Get(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("listing subscription groups: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response: %w", err)
		}

		var result Response[SubscriptionGroupResource]
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		for _, r := range result.Data {
			groups = append(groups, SubscriptionGroup{
				ID:            r.ID,
				ReferenceName: r.Attributes.ReferenceName,
			})
		}

		path = nextPath(result.Links.Next, c.BaseURL)
	}

	return groups, nil
}

// ListSubscriptions fetches subscriptions within a subscription group.
func ListSubscriptions(ctx context.Context, c *client.Client, groupID string) ([]SubscriptionInfo, error) {
	var subs []SubscriptionInfo
	path := fmt.Sprintf("/v1/subscriptionGroups/%s/subscriptions", groupID)

	for path != "" {
		resp, err := c.Get(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("listing subscriptions: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response: %w", err)
		}

		var result Response[SubscriptionResource]
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		for _, r := range result.Data {
			subs = append(subs, SubscriptionInfo{
				SubscriptionID:     r.ID,
				Name:               r.Attributes.Name,
				ProductID:          r.Attributes.ProductID,
				State:              r.Attributes.State,
				SubscriptionPeriod: r.Attributes.SubscriptionPeriod,
				GroupLevel:         r.Attributes.GroupLevel,
			})
		}

		path = nextPath(result.Links.Next, c.BaseURL)
	}

	return subs, nil
}

// FetchAppSubscriptions fetches all subscriptions for an app, grouped by subscription group.
// If params.GroupID is set, only that group's subscriptions are returned.
func FetchAppSubscriptions(ctx context.Context, c *client.Client, params SubscriptionParams) ([]SubscriptionInfo, error) {
	groups, err := ListSubscriptionGroups(ctx, c, params.AppID)
	if err != nil {
		return nil, err
	}

	var results []SubscriptionInfo
	for _, g := range groups {
		if params.GroupID != "" && g.ID != params.GroupID {
			continue
		}

		subs, err := ListSubscriptions(ctx, c, g.ID)
		if err != nil {
			return nil, fmt.Errorf("fetching subscriptions for group %s: %w", g.ID, err)
		}

		for i := range subs {
			subs[i].GroupID = g.ID
			subs[i].GroupName = g.ReferenceName
		}

		results = append(results, subs...)
	}

	return results, nil
}
