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

resource "sftpgo_defender_entry" "test" {
    ipornet    = "192.168.1.0/24"
    description = "created from Terraform"
    mode = 1
    protocols = 7
}
