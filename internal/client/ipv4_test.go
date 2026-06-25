package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestOrderServerRoutedLayer3IPv4(t *testing.T) {
	var posts int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/servers/17617" && posts == 0:
			_, _ = io.WriteString(w, `{"meta":{},"data":[{"srv_id":"17617","ips":[{"ip_id":"7795","ip_v4v6":"ipv4","ip_address":"185.181.63.24","ip_type":"primary"}]}]}`)
		case r.Method == http.MethodPost && r.URL.Path == "/servers/17617/ipv4":
			posts++
			var body orderServerIPv4Request
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.IPType != "l3" {
				t.Fatalf("ip_type = %q, want l3", body.IPType)
			}
			if body.NumIPs != "1" {
				t.Fatalf("num_ips = %q, want 1", body.NumIPs)
			}
			_, _ = io.WriteString(w, `{"meta":{"status":200},"data":[]}`)
		case r.Method == http.MethodGet && r.URL.Path == "/servers/17617" && posts == 1:
			_, _ = io.WriteString(w, `{"meta":{},"data":[{"srv_id":"17617","ips":[{"ip_id":"7795","ip_v4v6":"ipv4","ip_address":"185.181.63.24","ip_type":"primary"},{"ip_id":"7538","ip_v4v6":"ipv4","ip_address":"185.181.62.21","ip_type":"extra","ip_netmask":"255.255.255.0","ip_gateway":"185.181.62.1"}]}]}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	ip, err := c.OrderServerRoutedLayer3IPv4(context.Background(), "17617")
	if err != nil {
		t.Fatalf("OrderServerRoutedLayer3IPv4: %v", err)
	}
	if int64(ip.IPID) != 7538 || ip.IPAddress != "185.181.62.21" || ip.IPType != "extra" {
		t.Fatalf("ip = %+v", ip)
	}
}

func TestMoveServerIPv4(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/servers/17617/ipv4/7538" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body moveServerIPv4Request
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.TargetSrvID != 17618 {
			t.Fatalf("target_srv_id = %d, want 17618", body.TargetSrvID)
		}
		_, _ = io.WriteString(w, `{"meta":{"status":200,"message":"IP has been moved."},"data":{}}`)
	})

	if err := c.MoveServerIPv4(context.Background(), "17617", 7538, "17618"); err != nil {
		t.Fatalf("MoveServerIPv4: %v", err)
	}
}

func TestUpdateServerIPReverse(t *testing.T) {
	for _, tc := range []struct {
		name string
		v4v6 string
	}{
		{name: "IPv4", v4v6: "ipv4"},
		{name: "IPv6", v4v6: "ipv6"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut || r.URL.Path != "/servers/17617/reverse" {
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				var body updateServerReverseRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.IPID != 7538 || body.DNS != "server.example.com" || body.V4V6 != tc.v4v6 {
					t.Fatalf("body = %+v", body)
				}
				_, _ = io.WriteString(w, `{"meta":{"status":200,"message":"Reverse updated."},"data":{}}`)
			})

			if err := c.UpdateServerIPReverse(context.Background(), "17617", 7538, tc.v4v6, "server.example.com"); err != nil {
				t.Fatalf("UpdateServerIPReverse: %v", err)
			}
		})
	}
}

func TestDeleteServerIPv4(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/servers/17617/ipv4/7538" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"meta":{"status":200,"message":"IP removed."},"data":{}}`)
	})

	if err := c.DeleteServerIPv4(context.Background(), "17617", 7538); err != nil {
		t.Fatalf("DeleteServerIPv4: %v", err)
	}
}
