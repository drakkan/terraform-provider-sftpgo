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

resource "sftpgo_admin" "test" {
    username    = "test"
    status = 1
    password = "password"
    email = "admin@sftpgo.com"
    permissions = ["add_users", "edit_users","del_users"]
    filters = {
        allow_list = ["192.168.1.0/24"]
    }
    preferences = {
      hide_user_page_sections = 5
    }
}

output "sftpgo_admin" {
  value = sftpgo_admin.test
  sensitive = true
}