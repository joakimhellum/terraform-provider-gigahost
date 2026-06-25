package client

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
)

type orderServerIPv4Request struct {
	IPType string `json:"ip_type"`
	NumIPs string `json:"num_ips"`
}

type moveServerIPv4Request struct {
	TargetSrvID int64 `json:"target_srv_id"`
}

func (c *Client) OrderServerRoutedLayer3IPv4(ctx context.Context, srvID string) (*ServerIP, error) {
	if _, err := strconv.ParseInt(srvID, 10, 64); err != nil {
		return nil, fmt.Errorf("gigahost: invalid srv id %q", srvID)
	}
	before, err := c.GetServer(ctx, srvID)
	if err != nil {
		return nil, err
	}
	seen := map[int64]bool{}
	for _, ip := range before.IPs {
		if ip.IPv4v6 == "ipv4" && ip.IPType == "extra" {
			seen[int64(ip.IPID)] = true
		}
	}

	req, err := c.newRequest(ctx, http.MethodPost, path.Join("servers", srvID, "ipv4"), nil, orderServerIPv4Request{IPType: "l3", NumIPs: "1"})
	if err != nil {
		return nil, err
	}
	if err := c.sendRequest(req, nil); err != nil {
		return nil, err
	}

	after, err := c.GetServer(ctx, srvID)
	if err != nil {
		return nil, err
	}
	for _, ip := range after.IPs {
		if ip.IPv4v6 == "ipv4" && ip.IPType == "extra" && !seen[int64(ip.IPID)] {
			return &ip, nil
		}
	}
	return nil, fmt.Errorf("gigahost: ordered routed layer 3 IPv4 for server %s but could not find a new extra IPv4 on the server", srvID)
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
