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

resource "sftpgo_folder" "test" {
    name    = "test"
    mapped_path    = "/tmp/test1"
    filesystem = {
      provider = 0
    }
}

output "sftpgo_folder" {
  value = sftpgo_folder.test
  sensitive = true
}
