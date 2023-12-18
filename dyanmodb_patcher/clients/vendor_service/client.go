package vendorSrv

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/config"
	"github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/utils"
	pdkit "github.com/deliveryhero/pd-go-kit"
)

type Client struct {
	httpClient        *http.Client
	endpointFormatStr string
	serviceToken      string
	userEmail         string
	globalEntity      utils.GlobalEntity
}

type vendorServiceResp struct {
	AccountNameLocalized string `json:"account_name_localized"`
}

func NewClient(ge utils.GlobalEntity, cfg config.Config, httpClient *http.Client) *Client {
	var token = os.Getenv("VENDOR_SERVICE_TOKEN")
	var email = os.Getenv("EMAIL")

	return &Client{
		httpClient:        httpClient,
		endpointFormatStr: cfg.VendorService.EndpointFormatStr,
		serviceToken:      token,
		userEmail:         email,
		globalEntity:      ge,
	}
}

func (c *Client) ValidateEnvConfig() error {
	if c.serviceToken == "" {
		return fmt.Errorf("VENDOR_SERVICE_TOKEN env variable is required, please edit .env file")
	}

	if c.userEmail == "" {
		return fmt.Errorf("EMAIL env variable is required, please edit .env file")
	}
	return nil
}

func (c *Client) GetLocalLegalName(vendorCode string) (string, error) {
	req, err := http.NewRequest(
		http.MethodGet, fmt.Sprintf(c.endpointFormatStr, c.globalEntity.CountryCode, vendorCode), nil,
	)
	if err != nil {
		log.Fatalf("unable to make a new request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add(pdkit.HeaderAPIOAuthToken, "Bearer "+c.serviceToken)
	req.Header.Add(pdkit.HeaderAPIGlobalEntityID, c.globalEntity.ID)
	req.Header.Add("X-Pandora-Username", c.userEmail)
	req.Header.Add(pdkit.HeaderPerseusClientID, "no-user-interaction")
	req.Header.Add(pdkit.HeaderPerseusSessionID, "no-user-interaction")

	response, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to invoke vendor service: %w", err)
	}

	defer response.Body.Close()

	statusCode := response.StatusCode
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get vendor from vendor service: %d, %v", statusCode, response)
	}

	var resp vendorServiceResp
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("unable to decode response: %w", err)
	}

	// account_name_localized from vendor service = local legal name we pass to cybersource
	return resp.AccountNameLocalized, nil
}
