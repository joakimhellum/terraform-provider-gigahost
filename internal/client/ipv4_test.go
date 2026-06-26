package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestOrderServerIPv4(t *testing.T) {
	var posts int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/servers/1001/ipv4":
			posts++
			var body orderServerIPv4Request
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.IPType != "l2" {
				t.Fatalf("ip_type = %q, want l2", body.IPType)
			}
			if body.NumIPs != "1" {
				t.Fatalf("num_ips = %q, want 1", body.NumIPs)
			}
			_, _ = io.WriteString(w, `{"meta":{"status":200,"ip_id":"2002","ip_address":"192.0.2.20"},"data":[]}`)
		case r.Method == http.MethodGet && r.URL.Path == "/servers/1001" && posts == 1:
			_, _ = io.WriteString(w, `{"meta":{},"data":[{"srv_id":"1001","ips":[{"ip_id":"2001","ip_v4v6":"ipv4","ip_address":"192.0.2.10","ip_type":"primary"},{"ip_id":"2002","ip_v4v6":"ipv4","ip_address":"192.0.2.20","ip_type":"extra","ip_netmask":"255.255.255.0","ip_gateway":"192.0.2.1"}]}]}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	ip, err := c.OrderServerIPv4(context.Background(), "1001", "l2")
	if err != nil {
		t.Fatalf("OrderServerIPv4: %v", err)
	}
	if int64(ip.IPID) != 2002 || ip.IPAddress != "192.0.2.20" || ip.IPType != "extra" || ip.IPNetmask != "255.255.255.0" {
		t.Fatalf("ip = %+v", ip)
	}
}

func TestOrderServerIPv4UsesMetaWhenReadBackLags(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/servers/1001/ipv4":
			_, _ = io.WriteString(w, `{"meta":{"status":200,"ip_id":2002,"ip_address":"192.0.2.20"},"data":[]}`)
		case r.Method == http.MethodGet && r.URL.Path == "/servers/1001":
			_, _ = io.WriteString(w, `{"meta":{},"data":[{"srv_id":"1001","ips":[{"ip_id":"2001","ip_v4v6":"ipv4","ip_address":"192.0.2.10","ip_type":"primary"}]}]}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	ip, err := c.OrderServerIPv4(context.Background(), "1001", "l3")
	if err != nil {
		t.Fatalf("OrderServerIPv4: %v", err)
	}
	if int64(ip.IPID) != 2002 || ip.IPAddress != "192.0.2.20" || ip.IPv4v6 != "ipv4" || ip.IPType != "extra" {
		t.Fatalf("ip = %+v", ip)
	}
}

func TestMoveServerIPv4(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/servers/1001/ipv4/2002" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body moveServerIPv4Request
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.TargetSrvID != 1002 {
			t.Fatalf("target_srv_id = %d, want 1002", body.TargetSrvID)
		}
		_, _ = io.WriteString(w, `{"meta":{"status":200,"message":"IP has been moved."},"data":{}}`)
	})

	if err := c.MoveServerIPv4(context.Background(), "1001", 2002, "1002"); err != nil {
		t.Fatalf("MoveServerIPv4: %v", err)
	}
}

func TestDeleteServerIPv4(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/servers/1001/ipv4/2002" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"meta":{"status":200,"message":"IP removed."},"data":{}}`)
	})

	if err := c.DeleteServerIPv4(context.Background(), "1001", 2002); err != nil {
		t.Fatalf("DeleteServerIPv4: %v", err)
	}
}
