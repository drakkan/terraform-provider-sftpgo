terraform {
  required_providers {
    sftpgo = {
      source = "registry.terraform.io/drakkan/sftpgo"
    }
  }
}

provider "sftpgo" {
  host     = "http://localhost:8080"
  username = "admin"
  password = "password"
}

data "sftpgo_license_info" "current" {}

output "license_details" {
  value = data.sftpgo_license_info.current.license
}
