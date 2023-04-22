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

resource "sftpgo_rule" "test" {
    name = "backup"
    status = 1
    description = "created from Terraform"
    trigger = 1
    conditions = {
        fs_events = ["upload", "download"]
    }
    actions = [
        {
            name = "http action"
        }
    ]
}