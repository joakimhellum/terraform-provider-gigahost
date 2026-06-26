package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pigeon-as/terraform-provider-gigahost/internal/client"
)

func TestParseServerReverseImportID(t *testing.T) {
	serverID, v4v6, ipID, err := parseServerReverseImportID("1001/ipv4/2002")
	if err != nil {
		t.Fatalf("parseServerReverseImportID: %v", err)
	}
	if serverID != "1001" || v4v6 != "ipv4" || ipID != 2002 {
		t.Fatalf("parsed import id = %q, %q, %d", serverID, v4v6, ipID)
	}
}

func TestParseServerReverseImportIDRejectsInvalidIPVersion(t *testing.T) {
	_, _, _, err := parseServerReverseImportID("1001/ip4/2002")
	if err == nil {
		t.Fatal("expected invalid IP version to fail")
	}
}

func TestServerReverseStateFromIPPreservesExistingReverseWhenReadBackEmpty(t *testing.T) {
	state := serverReverseResourceModel{
		IPReverse: types.StringValue("server.example.com"),
	}
	ip := client.ServerIP{
		IPAddress: "2001:db8::1",
		IPType:    "primary",
		IPReverse: "",
	}

	newState := serverReverseStateFromIP(&state, ip, true)

	if got := newState.IPReverse.ValueString(); got != "server.example.com" {
		t.Fatalf("ip_reverse = %q", got)
	}
	if got := newState.IPAddress.ValueString(); got != "2001:db8::1" {
		t.Fatalf("ip_address = %q", got)
	}
	if got := newState.IPType.ValueString(); got != "primary" {
		t.Fatalf("ip_type = %q", got)
	}
}

func TestServerReverseStateFromIPUsesNonEmptyReadBackReverse(t *testing.T) {
	state := serverReverseResourceModel{
		IPReverse: types.StringValue("old.example.com"),
	}
	ip := client.ServerIP{IPReverse: "new.example.com"}

	newState := serverReverseStateFromIP(&state, ip, true)

	if got := newState.IPReverse.ValueString(); got != "new.example.com" {
		t.Fatalf("ip_reverse = %q", got)
	}
}

func TestServerReverseStateFromIPIgnoresReadBackReverseAfterWrite(t *testing.T) {
	state := serverReverseResourceModel{
		IPReverse: types.StringValue("planned.example.com"),
	}
	ip := client.ServerIP{IPReverse: "api.example.com"}

	newState := serverReverseStateFromIP(&state, ip, false)

	if got := newState.IPReverse.ValueString(); got != "planned.example.com" {
		t.Fatalf("ip_reverse = %q", got)
	}
}

func TestServerReverseStateFromIPDefaultsImportedEmptyReadBackReverse(t *testing.T) {
	state := serverReverseResourceModel{
		IPReverse: types.StringNull(),
	}
	ip := client.ServerIP{IPReverse: ""}

	newState := serverReverseStateFromIP(&state, ip, true)

	if got := newState.IPReverse.ValueString(); got != "" || newState.IPReverse.IsNull() {
		t.Fatalf("ip_reverse = %q null=%t", got, newState.IPReverse.IsNull())
	}
}
