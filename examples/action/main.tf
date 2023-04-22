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

resource "sftpgo_action" "test" {
    name = "http action"
    description = "created from Terraform"
    type = 1
    options = {
        http_config = {
            endpoint = "http://127.0.0.1:8082/notify"
            username = "myuser"
            password = "mypassword"
            timeout = 10
            method = "GET"
        }
    }
}

output "sftpgo_action" {
  value = sftpgo_action.test
  sensitive = true
}
