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

resource "sftpgo_user" "test" {
    username    = "test"
    status      = 1
    password    = "password"
    home_dir    = "/tmp/test1"
    email       = "test@test.com"
    permissions = {
        "/" = "*",
        "/p1" = "list,download"
    }
    filesystem = {
      provider = 1
      s3config = {
        bucket = "abc"
        region = "us-west-1"
        access_key = "key"
        access_secret = {
          status = "Plain"
          payload = "secret payload"
        }
      }
    }
    groups = [
      {
        name = "test"
        type = 3
      }
    ]
    virtual_folders = [
      {
        name = "test"
        virtual_path = "/vdir"
        quota_size = -1
        quota_files = -1
      }
    ]
    filters = {
      allowed_ip = ["192.168.1.0/24", "10.0.0.0/8"]
      start_directory = "/start/dir"
      file_patterns = [
        {
          path = "/p1"
          allowed_patterns = ["*.jpg","*.pdf"]
          deny_policy = 1
        },
        {
          path = "/p2"
          denied_patterns = ["*.jpg","*.pdf"]
        },
        {
          path = "/p3"
          denied_patterns = ["*.abc"]
        }
      ]
      hooks = {
        external_auth_disabled = true
      }
      bandwidth_limits = [
        {
          sources = ["127.0.0.1/32","192.168.1.0/24"]
          upload_bandwidth = 256
          download_bandwidth = 128
        }
      ]
    }
}

output "sftpgo_user" {
  value = sftpgo_user.test
  sensitive = true
}
