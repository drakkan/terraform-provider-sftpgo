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

resource "sftpgo_group" "test" {
    name    = "test"
    user_settings = {
        max_sessions = 10
        filters = {
          denied_protocols = ["FTP"]
          web_client = ["write-disabled", "password-change-disabled"]
        }
        filesystem = {
            provider = 0
        }
    }
    virtual_folders = [
      {
        name = "test"
        virtual_path = "/g1"
        quota_size = 0
        quota_files = 0
      }
    ]
}

output "sftpgo_group" {
  value = sftpgo_group.test
  sensitive = true
}