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
	_ resource.Resource                = &serverIPv4Resource{}
	_ resource.ResourceWithConfigure   = &serverIPv4Resource{}
	_ resource.ResourceWithImportState = &serverIPv4Resource{}
)

func NewServerIPv4Resource() resource.Resource {
	return &serverIPv4Resource{}
}

type serverIPv4Resource struct {
	client *client.Client
}

type serverIPv4ResourceModel struct {
	SrvId     types.String `tfsdk:"srv_id"`
	IPType    types.String `tfsdk:"ip_type"`
	IpId      types.Int64  `tfsdk:"ip_id"`
	IpAddress types.String `tfsdk:"ip_address"`
	IpReverse types.String `tfsdk:"ip_reverse"`
	IpNetmask types.String `tfsdk:"ip_netmask"`
	IpGateway types.String `tfsdk:"ip_gateway"`
}

func (r *serverIPv4Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_ipv4"
}

func (r *serverIPv4Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Orders and manages an extra IPv4 address on a Gigahost server. Layer 3 IPv4 addresses can be moved between servers.",
		Attributes: map[string]schema.Attribute{
			"srv_id": schema.StringAttribute{
				Required:            true,
				Description:         "Server id that currently holds the IPv4 address.",
				MarkdownDescription: "Server id that currently holds the IPv4 address.",
			},
			"ip_type": schema.StringAttribute{
				Required:            true,
				Description:         "IPv4 routing mode to order. Must be l2 or l3. The Gigahost read API reports only primary/extra, so this is kept from configuration and changing it replaces the resource.",
				MarkdownDescription: "IPv4 routing mode to order. Must be `l2` or `l3`. The Gigahost read API reports only `primary`/`extra`, so this is kept from configuration and changing it replaces the resource.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:          []validator.String{stringvalidator.OneOf("l2", "l3")},
			},
			"ip_id": schema.Int64Attribute{
				Computed:            true,
				Description:         "IP address id.",
				MarkdownDescription: "IP address id.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"ip_address": schema.StringAttribute{
				Computed:            true,
				Description:         "The IPv4 address.",
				MarkdownDescription: "The IPv4 address.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ip_reverse": schema.StringAttribute{
				Computed:            true,
				Description:         "Current reverse DNS (PTR) for the address.",
				MarkdownDescription: "Current reverse DNS (PTR) for the address.",
			},
			"ip_netmask": schema.StringAttribute{
				Computed:            true,
				Description:         "Netmask.",
				MarkdownDescription: "Netmask.",
			},
			"ip_gateway": schema.StringAttribute{
				Computed:            true,
				Description:         "Gateway.",
				MarkdownDescription: "Gateway.",
			},
		},
	}
}

func (r *serverIPv4Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serverIPv4Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverIPv4ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ip, err := r.client.OrderServerIPv4(ctx, plan.SrvId.ValueString(), plan.IPType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Order Gigahost Server IPv4", err.Error())
		return
	}

	state := plan
	serverIPv4StateFromAPI(&state, plan.SrvId.ValueString(), ip)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serverIPv4Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverIPv4ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ip, serverID, err := r.findIPv4(ctx, state.SrvId.ValueString(), state.IpId.ValueInt64())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read Gigahost Server IPv4", err.Error())
		return
	}

	serverIPv4StateFromAPI(&state, serverID, ip)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serverIPv4Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serverIPv4ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.SrvId.ValueString() != state.SrvId.ValueString() {
		if err := r.client.MoveServerIPv4(ctx, state.SrvId.ValueString(), state.IpId.ValueInt64(), plan.SrvId.ValueString()); err != nil {
			resp.Diagnostics.AddError("Unable to Move Gigahost Server IPv4", err.Error())
			return
		}
	}
	ip, serverID, err := r.findIPv4(ctx, plan.SrvId.ValueString(), state.IpId.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Gigahost Server IPv4 After Update", err.Error())
		return
	}

	newState := plan
	serverIPv4StateFromAPI(&newState, serverID, ip)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *serverIPv4Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverIPv4ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteServerIPv4(ctx, state.SrvId.ValueString(), state.IpId.ValueInt64()); err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to Delete Gigahost Server IPv4", err.Error())
	}
}

func (r *serverIPv4Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import Id",
			fmt.Sprintf("Expected import id in the format \"srv_id/ip_id/ip_type\", got: %q.", req.ID),
		)
		return
	}
	serverID, ipID, ipType := parts[0], parts[1], parts[2]
	if ipType != "l2" && ipType != "l3" {
		resp.Diagnostics.AddError("Invalid Import Id", fmt.Sprintf("IP type must be l2 or l3, got %q.", ipType))
		return
	}
	parsedIPID, err := strconv.ParseInt(ipID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import Id", fmt.Sprintf("IP id %q is not numeric: %v.", ipID, err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("srv_id"), serverID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip_id"), parsedIPID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip_type"), ipType)...)
}

func (r *serverIPv4Resource) findIPv4(ctx context.Context, serverID string, ipID int64) (*client.ServerIP, string, error) {
	server, err := r.client.GetServer(ctx, serverID)
	if err == nil {
		if ip := findServerIPv4(server.IPs, ipID); ip != nil {
			return ip, serverID, nil
		}
	}
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		return nil, "", err
	}

	servers, err := r.client.ListServers(ctx)
	if err != nil {
		return nil, "", err
	}
	for _, server := range servers {
		if ip := findServerIPv4(server.IPs, ipID); ip != nil {
			return ip, server.SrvID, nil
		}
	}
	return nil, "", client.ErrNotFound
}

func findServerIPv4(ips []client.ServerIP, ipID int64) *client.ServerIP {
	for _, ip := range ips {
		if int64(ip.IPID) == ipID && ip.IPv4v6 == "ipv4" && ip.IPType == "extra" {
			return &ip
		}
	}
	return nil
}

func serverIPv4StateFromAPI(state *serverIPv4ResourceModel, serverID string, ip *client.ServerIP) {
	state.SrvId = types.StringValue(serverID)
	state.IpId = types.Int64Value(int64(ip.IPID))
	state.IpAddress = types.StringValue(ip.IPAddress)
	state.IpReverse = types.StringValue(ip.IPReverse)
	state.IpNetmask = types.StringValue(ip.IPNetmask)
	state.IpGateway = types.StringValue(ip.IPGateway)
}
