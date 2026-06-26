resource "gigahost_server_reverse" "example" {
  srv_id     = gigahost_server.example.srv_id
  ip_id      = 2002
  ip_v4v6    = "ipv4"
  ip_reverse = "server.example.com"
}
