package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestUpdateServerReverse(t *testing.T) {
	for _, tc := range []struct {
		name string
		v4v6 string
	}{
		{name: "IPv4", v4v6: "ipv4"},
		{name: "IPv6", v4v6: "ipv6"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut || r.URL.Path != "/servers/1001/reverse" {
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				var body updateServerReverseRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.IPID != 2002 || body.DNS != "server.example.com" || body.V4V6 != tc.v4v6 {
					t.Fatalf("body = %+v", body)
				}
				_, _ = io.WriteString(w, `{"meta":{"status":200,"message":"Reverse updated."},"data":{}}`)
			})

			if err := c.UpdateServerReverse(context.Background(), "1001", 2002, tc.v4v6, "server.example.com"); err != nil {
				t.Fatalf("UpdateServerReverse: %v", err)
			}
		})
	}
}
