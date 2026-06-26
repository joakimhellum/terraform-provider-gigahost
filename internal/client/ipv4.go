package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"

	"github.com/hashicorp/go-retryablehttp"
)

type orderServerIPv4Request struct {
	IPType string `json:"ip_type"`
	NumIPs string `json:"num_ips"`
}

type orderServerIPv4Meta struct {
	IPID      json.RawMessage `json:"ip_id"`
	IPAddress string          `json:"ip_address"`
}

type orderServerIPv4Response struct {
	Meta orderServerIPv4Meta `json:"meta"`
}

type moveServerIPv4Request struct {
	TargetSrvID int64 `json:"target_srv_id"`
}

func (c *Client) OrderServerIPv4(ctx context.Context, srvID string, ipType string) (*ServerIP, error) {
	if _, err := strconv.ParseInt(srvID, 10, 64); err != nil {
		return nil, fmt.Errorf("gigahost: invalid srv id %q", srvID)
	}
	if ipType != "l2" && ipType != "l3" {
		return nil, fmt.Errorf("gigahost: invalid IPv4 ip_type %q", ipType)
	}

	req, err := c.newRequest(ctx, http.MethodPost, path.Join("servers", srvID, "ipv4"), nil, orderServerIPv4Request{IPType: ipType, NumIPs: "1"})
	if err != nil {
		return nil, err
	}
	ordered, err := c.sendOrderServerIPv4Request(req)
	if err != nil {
		return nil, err
	}

	server, err := c.GetServer(ctx, srvID)
	if err == nil {
		for _, ip := range server.IPs {
			if ip.IPID == ordered.IPID && ip.IPv4v6 == "ipv4" && ip.IPType == "extra" {
				return &ip, nil
			}
		}
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	return ordered, nil
}

func (c *Client) sendOrderServerIPv4Request(req *retryablehttp.Request) (*ServerIP, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing %s %s: %w", req.Method, req.URL.Path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	if err := checkResponse(resp.StatusCode, body); err != nil {
		return nil, err
	}

	var out orderServerIPv4Response
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decoding response envelope: %w", err)
	}
	ipID, err := parseOrderServerIPv4ID(out.Meta.IPID)
	if err != nil {
		return nil, err
	}
	if out.Meta.IPAddress == "" {
		return nil, fmt.Errorf("gigahost: order server IPv4 response missing meta.ip_address")
	}
	return &ServerIP{IPID: flexInt64(ipID), IPAddress: out.Meta.IPAddress, IPv4v6: "ipv4", IPType: "extra"}, nil
}

func parseOrderServerIPv4ID(raw json.RawMessage) (int, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, fmt.Errorf("gigahost: order server IPv4 response missing meta.ip_id")
	}
	var idInt int
	if err := json.Unmarshal(raw, &idInt); err == nil {
		if idInt <= 0 {
			return 0, fmt.Errorf("gigahost: invalid ordered IPv4 id %d", idInt)
		}
		return idInt, nil
	}
	var idString string
	if err := json.Unmarshal(raw, &idString); err != nil {
		return 0, fmt.Errorf("gigahost: decoding ordered IPv4 id: %w", err)
	}
	parsed, err := strconv.Atoi(idString)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("gigahost: invalid ordered IPv4 id %q", idString)
	}
	return parsed, nil
}

func (c *Client) MoveServerIPv4(ctx context.Context, sourceSrvID string, ipID int64, targetSrvID string) error {
	if _, err := strconv.ParseInt(sourceSrvID, 10, 64); err != nil {
		return fmt.Errorf("gigahost: invalid source srv id %q", sourceSrvID)
	}
	targetID, err := strconv.ParseInt(targetSrvID, 10, 64)
	if err != nil {
		return fmt.Errorf("gigahost: invalid target srv id %q", targetSrvID)
	}
	if ipID <= 0 {
		return fmt.Errorf("gigahost: invalid ip id %d", ipID)
	}

	req, err := c.newRequest(ctx, http.MethodPut, path.Join("servers", sourceSrvID, "ipv4", strconv.FormatInt(ipID, 10)), nil, moveServerIPv4Request{TargetSrvID: targetID})
	if err != nil {
		return err
	}
	return c.sendRequest(req, nil)
}

func (c *Client) DeleteServerIPv4(ctx context.Context, srvID string, ipID int64) error {
	if _, err := strconv.ParseInt(srvID, 10, 64); err != nil {
		return fmt.Errorf("gigahost: invalid srv id %q", srvID)
	}
	if ipID <= 0 {
		return fmt.Errorf("gigahost: invalid ip id %d", ipID)
	}

	req, err := c.newRequest(ctx, http.MethodDelete, path.Join("servers", srvID, "ipv4", strconv.FormatInt(ipID, 10)), nil, nil)
	if err != nil {
		return err
	}
	return c.sendRequest(req, nil)
}
