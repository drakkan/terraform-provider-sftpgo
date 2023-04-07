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
      provider = 3
      azblobconfig = {
        container = "fake container"
        account_name = "my access key"
        account_key = "my secret"
        key_prefix = "prefix/"
      }
    }
}

output "sftpgo_folder" {
  value = sftpgo_folder.test
  sensitive = true
}
