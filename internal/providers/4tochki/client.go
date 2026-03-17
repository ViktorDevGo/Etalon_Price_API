package fourtochki

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	soapEnvelopeNS = "http://schemas.xmlsoap.org/soap/envelope/"
	targetNS       = "Wcf.ClientService.Client.WebAPI.TS3"
	arrayNS        = "http://schemas.microsoft.com/2003/10/Serialization/Arrays"
	ts3NS          = "http://schemas.datacontract.org/2004/07/TS3.Domain.Models.Client.ClientSoapService.SearchProductsAndPricesInfo"
)

// SOAPEnvelope represents SOAP envelope structure
type SOAPEnvelope struct {
	XMLName   xml.Name `xml:"soapenv:Envelope"`
	SoapenvNS string   `xml:"xmlns:soapenv,attr"`
	WcfNS     string   `xml:"xmlns:wcf,attr"`
	ArrNS     string   `xml:"xmlns:arr,attr"`
	TS3NS     string   `xml:"xmlns:ts3,attr,omitempty"`
	Body      SOAPBody `xml:"soapenv:Body"`
}

// SOAPBody represents SOAP body
type SOAPBody struct {
	Content interface{} `xml:",omitempty"`
	Fault   *SOAPFault  `xml:"Fault,omitempty"`
}

// SOAPFault represents SOAP fault
type SOAPFault struct {
	Code   string `xml:"faultcode"`
	String string `xml:"faultstring"`
	Detail string `xml:"detail,omitempty"`
}

func (f *SOAPFault) Error() string {
	return fmt.Sprintf("SOAP Fault: %s - %s (Detail: %s)", f.Code, f.String, f.Detail)
}

// Client is a SOAP client for 4tochki API
type Client struct {
	httpClient  *http.Client
	wsdlURL     string
	login       string
	password    string
	retryCount  int
	retryDelay  time.Duration
	logger      *slog.Logger
}

// ClientConfig holds client configuration
type ClientConfig struct {
	WSDLURL    string
	Login      string
	Password   string
	Timeout    time.Duration
	RetryCount int
	RetryDelay time.Duration
	Logger     *slog.Logger
}

// NewClient creates a new SOAP client
func NewClient(cfg ClientConfig) *Client {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		wsdlURL:    cfg.WSDLURL,
		login:      cfg.Login,
		password:   cfg.Password,
		retryCount: cfg.RetryCount,
		retryDelay: cfg.RetryDelay,
		logger:     cfg.Logger,
	}
}

// call executes a SOAP request with retry logic
func (c *Client) call(ctx context.Context, action string, request, response interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.retryCount; attempt++ {
		if attempt > 0 {
			c.logger.Info("Retrying SOAP request",
				"action", action,
				"attempt", attempt,
				"max_retries", c.retryCount,
			)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.retryDelay * time.Duration(attempt)):
			}
		}

		err := c.doCall(ctx, action, request, response)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on SOAP faults or context errors
		if _, isFault := err.(*SOAPFault); isFault {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		c.logger.Warn("SOAP request failed",
			"action", action,
			"attempt", attempt,
			"error", err,
		)
	}

	return fmt.Errorf("SOAP request failed after %d retries: %w", c.retryCount, lastErr)
}

// doCall performs a single SOAP request
func (c *Client) doCall(ctx context.Context, action string, request, response interface{}) error {
	// Build SOAP envelope
	envelope := SOAPEnvelope{
		SoapenvNS: soapEnvelopeNS,
		WcfNS:     targetNS,
		ArrNS:     arrayNS,
		TS3NS:     ts3NS,
		Body: SOAPBody{
			Content: request,
		},
	}

	// Marshal request
	reqBody, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SOAP request: %w", err)
	}

	// Add XML declaration
	fullReq := append([]byte(xml.Header), reqBody...)

	c.logger.Debug("SOAP request",
		"action", action,
		"url", c.wsdlURL,
		"body_size", len(fullReq),
		"body", string(fullReq),
	)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.wsdlURL, bytes.NewReader(fullReq))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "text/xml; charset=utf-8")
	httpReq.Header.Set("SOAPAction", fmt.Sprintf("%s/ClientService/%s", targetNS, action))

	// Execute request
	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("HTTP request failed",
			"action", action,
			"duration", duration,
			"error", err,
		)
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Debug("SOAP response",
		"action", action,
		"status", httpResp.StatusCode,
		"duration", duration,
		"body_size", len(respBody),
		"body", string(respBody),
	)

	// Check HTTP status
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: status %d, body: %s", httpResp.StatusCode, string(respBody))
	}

	// Parse response envelope - first check for faults
	var faultEnvelope struct {
		Body struct {
			Fault *SOAPFault `xml:"Fault"`
		} `xml:"Body"`
	}

	if err := xml.Unmarshal(respBody, &faultEnvelope); err == nil && faultEnvelope.Body.Fault != nil {
		c.logger.Error("SOAP fault received",
			"action", action,
			"fault_code", faultEnvelope.Body.Fault.Code,
			"fault_string", faultEnvelope.Body.Fault.String,
		)
		return faultEnvelope.Body.Fault
	}

	// Parse actual response - use flexible struct
	var respEnvelope struct {
		Body struct {
			Content interface{} `xml:",any"`
		} `xml:"Body"`
	}
	respEnvelope.Body.Content = response

	if err := xml.Unmarshal(respBody, &respEnvelope); err != nil {
		return fmt.Errorf("failed to unmarshal SOAP response: %w (body: %s)", err, string(respBody))
	}

	c.logger.Info("SOAP request successful",
		"action", action,
		"duration", duration,
	)

	return nil
}

// GetGoodsInfo retrieves goods information
func (c *Client) GetGoodsInfo(ctx context.Context, codes []string) (*GetGoodsInfoResponse, error) {
	req := &GetGoodsInfoRequest{
		Login:    c.login,
		Password: c.password,
		CodeList: StringArray{
			Items: codes,
		},
	}

	resp := &GetGoodsInfoResponse{}

	if err := c.call(ctx, "GetGoodsInfo", req, resp); err != nil {
		return nil, fmt.Errorf("GetGoodsInfo failed: %w", err)
	}

	return resp, nil
}

// GetGoodsPriceRestByCode retrieves prices and stock by codes
func (c *Client) GetGoodsPriceRestByCode(ctx context.Context, codes []string) (*GetGoodsPriceRestByCodeResponse, error) {
	req := &GetGoodsPriceRestByCodeRequest{
		Login:    c.login,
		Password: c.password,
		Filter: PriceRestFilter{
			CodeList: StringArrayTS3{
				Items: codes,
			},
			IncludePaidDelivery: true,
		},
	}

	resp := &GetGoodsPriceRestByCodeResponse{}

	if err := c.call(ctx, "GetGoodsPriceRestByCode", req, resp); err != nil {
		return nil, fmt.Errorf("GetGoodsPriceRestByCode failed: %w", err)
	}

	return resp, nil
}

// GetFindTyre retrieves all available tyre codes with pagination
func (c *Client) GetFindTyre(ctx context.Context, page int, pageSize int) (*GetFindTyreResponse, error) {
	if pageSize <= 0 {
		pageSize = 2000 // Default page size
	}
	if page < 0 {
		page = 0
	}

	req := &GetFindTyreRequest{
		Login:    c.login,
		Password: c.password,
		Filter:   EmptyFilter{},
		Page:     page,
		PageSize: pageSize,
	}

	resp := &GetFindTyreResponse{}

	if err := c.call(ctx, "GetFindTyre", req, resp); err != nil {
		return nil, fmt.Errorf("GetFindTyre failed: %w", err)
	}

	return resp, nil
}

// GetFindDisk retrieves all available disk codes with pagination
func (c *Client) GetFindDisk(ctx context.Context, page int, pageSize int) (*GetFindDiskResponse, error) {
	if pageSize <= 0 {
		pageSize = 2000 // Default page size
	}
	if page < 0 {
		page = 0
	}

	req := &GetFindDiskRequest{
		Login:    c.login,
		Password: c.password,
		Filter:   EmptyFilter{},
		Page:     page,
		PageSize: pageSize,
	}

	resp := &GetFindDiskResponse{}

	if err := c.call(ctx, "GetFindDisk", req, resp); err != nil {
		return nil, fmt.Errorf("GetFindDisk failed: %w", err)
	}

	return resp, nil
}

// GetWarehouses retrieves list of all warehouses
func (c *Client) GetWarehouses(ctx context.Context) (*GetWarehousesResponse, error) {
	req := &GetWarehousesRequest{
		Login:    c.login,
		Password: c.password,
	}

	resp := &GetWarehousesResponse{}

	if err := c.call(ctx, "GetWarehouses", req, resp); err != nil {
		return nil, fmt.Errorf("GetWarehouses failed: %w", err)
	}

	return resp, nil
}
