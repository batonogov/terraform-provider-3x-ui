data "3xui_server_status" "current" {}

output "current_cpu" {
  value = data.3xui_server_status.current.cpu.current
}
