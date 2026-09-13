# Example: Creating a virtual folder resource first, then referencing in a group
# This approach creates the folder as a separate resource, then uses it in the group.

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

# Step 1: Create the virtual folder as a separate resource
resource "sftpgo_folder" "shared_folder" {
  name        = "shared-data"
  mapped_path = "/srv/sftpgo/shared"
  description = "Shared data folder for multiple groups"
  filesystem = {
    provider = 0  # Local filesystem
  }
}

# Step 2: Create a group that uses the folder by name
# The virtual_folders block only needs:
# - name: matches the sftpgo_folder resource name
# - virtual_path: where to mount this folder for this group
# - quota_size: optional quota for this group
# - quota_files: optional file count quota
resource "sftpgo_group" "with_folder_resource" {
  name    = "folder-ref-group"
  user_settings = {
    max_sessions = 10
    filesystem = {
      provider = 0
    }
  }
  virtual_folders = [
    {
      name        = sftpgo_folder.shared_folder.name
      virtual_path = "/shared"
      quota_size   = 10737418240  # 10 GB
      quota_files  = 0
    }
  ]
}

output "sftpgo_folder" {
  value = sftpgo_folder.shared_folder
  sensitive = true
}

output "sftpgo_group_folder_ref" {
  value = sftpgo_group.with_folder_resource
  sensitive = true
}