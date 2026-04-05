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

resource "sftpgo_license" "test" {
  # key    = "1212-1212-1212-1212"
  # or use the write-only attribute for ephemeral values
  key_wo         = "1212-1212-1212-1212"
  key_wo_version = 1
}
