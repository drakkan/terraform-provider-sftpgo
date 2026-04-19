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
  edition = 1
}

variable "private_key" {
   type = string
   default = <<EOF
-----BEGIN PGP PRIVATE KEY BLOCK-----

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
=2nRD
-----END PGP PRIVATE KEY BLOCK-----
EOF
}

variable "public_key" {
   type = string
   default = <<EOF
-----BEGIN PGP PUBLIC KEY BLOCK-----

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
GRw45d4+RU0GqiutTB/J5RfVUzAbXAvG
=OvEA
-----END PGP PUBLIC KEY BLOCK-----
EOF
}

resource "sftpgo_action" "test" {
    name = "pgp action"
    description = "created from Terraform"
    type = 9
    options = {
      fs_config = {
         type = 7
         pgp = {
           mode = 2
           paths = [
             {
               key = "/{{.VirtualPath}}"
               value = "/{{.VirtualPath}}.pgp"
             }
           ],
           passphrase_wo          = "password",
           passphrase_wo_version  = "1",
           private_key_wo         = var.private_key,
           private_key_wo_version = "1",
           public_key             = var.public_key
         },
         "folder": "aaa",
         "target_folder": "bbbb"
      }
    }
}

output "sftpgo_action" {
  value = sftpgo_action.test
  sensitive = true
}
