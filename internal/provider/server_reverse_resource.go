package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pigeon-as/terraform-provider-gigahost/internal/client"
)

var (
	_ resource.Resource                = &serverReverseResource{}
	_ resource.ResourceWithConfigure   = &serverReverseResource{}
	_ resource.ResourceWithImportState = &serverReverseResource{}
)

func NewServerReverseResource() resource.Resource {
	return &serverReverseResource{}
}

type serverReverseResource struct {
	client *client.Client
}

type serverReverseResourceModel struct {
	SrvID     types.String `tfsdk:"srv_id"`
	IPID      types.Int64  `tfsdk:"ip_id"`
	IPv4v6    types.String `tfsdk:"ip_v4v6"`
	IPAddress types.String `tfsdk:"ip_address"`
	IPType    types.String `tfsdk:"ip_type"`
	IPReverse types.String `tfsdk:"ip_reverse"`
}

func (r *serverReverseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_reverse"
}

func (r *serverReverseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages reverse DNS (PTR) for a Gigahost server address.",
		Attributes: map[string]schema.Attribute{
			"srv_id": schema.StringAttribute{
				Required:            true,
				Description:         "Server id that holds the IP address.",
				MarkdownDescription: "Server id that holds the IP address.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ip_id": schema.Int64Attribute{
				Required:            true,
				Description:         "IP address id.",
				MarkdownDescription: "IP address id.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"ip_v4v6": schema.StringAttribute{
				Required:            true,
				Description:         "IP version for the address. Valid values are `ipv4` and `ipv6`.",
				MarkdownDescription: "IP version for the address. Valid values are `ipv4` and `ipv6`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:          []validator.String{stringvalidator.OneOf("ipv4", "ipv6")},
			},
			"ip_reverse": schema.StringAttribute{
				Required:            true,
				Description:         "Reverse DNS (PTR) for the address. Use an empty string to clear it.",
				MarkdownDescription: "Reverse DNS (PTR) for the address. Use an empty string to clear it.",
			},
			"ip_address": schema.StringAttribute{
				Computed:            true,
				Description:         "The IP address.",
				MarkdownDescription: "The IP address.",
			},
			"ip_type": schema.StringAttribute{
				Computed:            true,
				Description:         "IP address type, for example `primary` or `extra`.",
				MarkdownDescription: "IP address type, for example `primary` or `extra`.",
			},
		},
	}
}

func (r *serverReverseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *serverReverseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverReverseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.updateReverse(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to Update Gigahost Server Reverse DNS", err.Error())
		return
	}

	state, err := r.readState(ctx, &plan, false)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Gigahost Server Reverse DNS", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *serverReverseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverReverseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.readState(ctx, &state, true)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read Gigahost Server Reverse DNS", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *serverReverseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverReverseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.updateReverse(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to Update Gigahost Server Reverse DNS", err.Error())
		return
	}

	state, err := r.readState(ctx, &plan, false)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Gigahost Server Reverse DNS After Update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *serverReverseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverReverseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.IPReverse = types.StringValue("")
	if err := r.updateReverse(ctx, &state); err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to Clear Gigahost Server Reverse DNS", err.Error())
	}
}

func (r *serverReverseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	serverID, v4v6, ipID, err := parseServerReverseImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import Id", fmt.Sprintf("Expected import id in the format \"srv_id/ip_v4v6/ip_id\", got: %q.", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("srv_id"), serverID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip_v4v6"), v4v6)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip_id"), ipID)...)
}

func parseServerReverseImportID(id string) (string, string, int64, error) {
	serverID, rest, ok := strings.Cut(id, "/")
	if !ok {
		return "", "", 0, fmt.Errorf("invalid import id %q", id)
	}
	v4v6, ipID, ok := strings.Cut(rest, "/")
	if !ok || serverID == "" || v4v6 == "" || ipID == "" {
		return "", "", 0, fmt.Errorf("invalid import id %q", id)
	}
	if v4v6 != "ipv4" && v4v6 != "ipv6" {
		return "", "", 0, fmt.Errorf("invalid IP version %q", v4v6)
	}
	parsedIPID, err := strconv.ParseInt(ipID, 10, 64)
	if err != nil {
		return "", "", 0, fmt.Errorf("ip id %q is not numeric: %w", ipID, err)
	}
	return serverID, v4v6, parsedIPID, nil
}

func (r *serverReverseResource) updateReverse(ctx context.Context, state *serverReverseResourceModel) error {
	v4v6 := state.IPv4v6.ValueString()
	if v4v6 != "ipv4" && v4v6 != "ipv6" {
		return fmt.Errorf("gigahost: invalid IP version %q", v4v6)
	}
	return r.client.UpdateServerReverse(ctx, state.SrvID.ValueString(), state.IPID.ValueInt64(), v4v6, state.IPReverse.ValueString())
}

func (r *serverReverseResource) readState(ctx context.Context, state *serverReverseResourceModel, useAPIReverse bool) (*serverReverseResourceModel, error) {
	server, err := r.client.GetServer(ctx, state.SrvID.ValueString())
	if err != nil {
		return nil, err
	}

	ip := findServerIP(server.IPs, state.IPID.ValueInt64(), state.IPv4v6.ValueString())
	if ip == nil {
		return nil, client.ErrNotFound
	}

	newState := serverReverseStateFromIP(state, *ip, useAPIReverse)
	return &newState, nil
}

func serverReverseStateFromIP(state *serverReverseResourceModel, ip client.ServerIP, useAPIReverse bool) serverReverseResourceModel {
	newState := *state
	newState.IPAddress = types.StringValue(ip.IPAddress)
	newState.IPType = types.StringValue(ip.IPType)
	if useAPIReverse && ip.IPReverse != "" {
		newState.IPReverse = types.StringValue(ip.IPReverse)
	} else if newState.IPReverse.IsNull() || newState.IPReverse.IsUnknown() {
		newState.IPReverse = types.StringValue("")
	}
	return newState
}

func findServerIP(ips []client.ServerIP, ipID int64, v4v6 string) *client.ServerIP {
	for _, ip := range ips {
		if int64(ip.IPID) == ipID && ip.IPv4v6 == v4v6 {
			return &ip
		}
	}
	return nil
}
