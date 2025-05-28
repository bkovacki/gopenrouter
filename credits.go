package gopenrouter

import (
	"context"
	"net/http"
)

// creditsResponse represents the internal API response structure when retrieving credits information.
// It wraps the credits data in a standard response structure.
type creditsResponse struct {
	Data CreditsData `json:"data"`
}

// CreditsData contains information about a user's credits and usage.
// This provides visibility into the account's financial standing with OpenRouter.
type CreditsData struct {
	// TotalCredits represents the total amount of credits purchased or added to the account
	TotalCredits float64 `json:"total_credits"`
	// TotalUsage represents the total amount of credits consumed by API requests
	TotalUsage float64 `json:"total_usage"`
}

// GetCredits retrieves information about the authenticated user's credits and usage.
//
// This method provides a way to check the account's financial status, including
// the total purchased credits and how much has been consumed. This can be used
// for budgeting, monitoring usage, or determining when to purchase more credits.
//
// Parameters:
//   - ctx: The context for the request, which can be used for cancellation and timeouts
//
// Returns:
//   - CreditsData: Contains information about credits and usage
//   - error: Any error that occurred during the request
func (c *Client) GetCredits(ctx context.Context) (data CreditsData, err error) {
	urlSuffix := "/credits"
	var response creditsResponse

	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	if err != nil {
		return
	}

	data = response.Data
	return
}
