package client

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
)

type updateServerReverseRequest struct {
	IPID int64  `json:"ip_id"`
	DNS  string `json:"dns"`
	V4V6 string `json:"v4v6"`
}

func (c *Client) UpdateServerReverse(ctx context.Context, srvID string, ipID int64, v4v6 string, dns string) error {
	if _, err := strconv.ParseInt(srvID, 10, 64); err != nil {
		return fmt.Errorf("gigahost: invalid srv id %q", srvID)
	}
	if ipID <= 0 {
		return fmt.Errorf("gigahost: invalid ip id %d", ipID)
	}
	if v4v6 != "ipv4" && v4v6 != "ipv6" {
		return fmt.Errorf("gigahost: invalid IP version %q", v4v6)
	}

	req, err := c.newRequest(ctx, http.MethodPut, path.Join("servers", srvID, "reverse"), nil, updateServerReverseRequest{IPID: ipID, DNS: dns, V4V6: v4v6})
	if err != nil {
		return err
	}
	return c.sendRequest(req, nil)
}
