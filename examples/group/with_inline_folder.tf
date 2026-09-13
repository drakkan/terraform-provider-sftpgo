# Example: Creating a virtual folder inline within a group
# When using inline virtual_folders, you MUST provide mapped_path and filesystem.
# Without these fields, SFTPGo will fail with:
# "NOT NULL constraint failed: groups_folders_mapping.folder_id"

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

resource "sftpgo_group" "inline_example" {
  name    = "inline-folder-group"
  user_settings = {
    max_sessions = 10
    filesystem = {
      provider = 0
    }
  }
  # Required fields for inline virtual folders:
  # - name: unique folder identifier
  # - virtual_path: mount point within the group
  # - quota_size: max size in bytes (0 = unlimited)
  # - quota_files: max number of files (0 = unlimited)
  # - mapped_path: absolute path on the filesystem (REQUIRED!)
  # - filesystem: storage provider configuration (REQUIRED!)
  virtual_folders = [
    {
      name        = "data-folder"
      virtual_path = "/data"
      quota_size   = 0
      quota_files  = 0
      mapped_path  = "/srv/sftpgo/data"
      filesystem = {
        provider = 0  # Local filesystem
      }
    }
  ]
}

output "sftpgo_group_inline" {
  value = sftpgo_group.inline_example
  sensitive = true
}