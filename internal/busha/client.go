package busha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL		string
	bearerToken	string
	profileID	string
	httpClient	*http.Client
}

func NewClient(apiKey, baseURL, profileID string) *Client {
	return &Client{
		bearerToken:	apiKey,
		baseURL:	strings.TrimRight(baseURL, "/"),
		profileID:	profileID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, body, result interface{}) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
		slog.Debug("busha request", "method", method, "url", url, "body", string(b))
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.profileID != "" {
		req.Header.Set("X-BU-PROFILE-ID", c.profileID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	slog.Debug("busha response", "status", resp.StatusCode, "body", string(respBody))

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Error.Name != "" {
			msg := fmt.Sprintf("busha API error (%d): %s - %s", resp.StatusCode, apiErr.Error.Name, apiErr.Error.Message)
			if len(apiErr.Fields) > 0 {
				for field, reasons := range apiErr.Fields {
					msg += fmt.Sprintf(" | %s: %v", field, reasons)
				}
			}
			return fmt.Errorf(msg)
		}
		return fmt.Errorf("busha API error (%d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) CreateQuote(currency, targetAmount, recipientAddr, network string) (*Quote, error) {
	req := CreateQuoteRequest{
		SourceCurrency:	currency,
		TargetCurrency:	currency,
		TargetAmount:	targetAmount,
	}

	if recipientAddr != "" {
		req.PayOut = &PayOut{
			Type:		PayOutAddress,
			Address:	recipientAddr,
			Network:	network,
		}
	}

	var resp QuoteResponse
	if err := c.doRequest("POST", "/v1/quotes", req, &resp); err != nil {
		return nil, fmt.Errorf("create quote: %w", err)
	}

	slog.Info("quote created",
		"quote_id", resp.Data.ID,
		"source_amount", resp.Data.SourceAmount,
		"target_amount", resp.Data.TargetAmount,
		"status", resp.Data.Status,
	)

	return &resp.Data, nil
}

func (c *Client) CreateTransfer(quoteID string) (*Transfer, error) {
	req := map[string]string{"quote_id": quoteID}

	var resp TransferResponse
	if err := c.doRequest("POST", "/v1/transfers", req, &resp); err != nil {
		return nil, fmt.Errorf("create transfer: %w", err)
	}

	slog.Info("transfer created",
		"transfer_id", resp.Data.ID,
		"quote_id", resp.Data.QuoteID,
		"status", resp.Data.Status,
	)

	return &resp.Data, nil
}
