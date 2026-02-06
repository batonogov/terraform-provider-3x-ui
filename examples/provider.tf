terraform {
  required_providers {
    threexui = {
      source = "hashicorp/threexui"
    }
  }
}

provider "threexui" {
  endpoint             = "http://localhost:2053"
  base_path            = "/"
  username             = "admin"
  password             = "admin"
  insecure_skip_verify = true
  request_timeout      = "30s"
}
