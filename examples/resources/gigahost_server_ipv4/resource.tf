resource "gigahost_server_ipv4" "l2" {
  srv_id  = gigahost_server.example.srv_id
  ip_type = "l2"
}

resource "gigahost_server_ipv4" "l3" {
  srv_id  = gigahost_server.example.srv_id
  ip_type = "l3"
}
