terraform {
  required_providers {
    sftpgo = {
      source = "registry.terraform.io/drakkan/sftpgo"
    }
  }
}

data "google_service_account_id_token" "sftp_iap_oidc" {
  target_service_account = var.sftp_service_account
  target_audience        = var.sftp_iap_client_id
  include_email          = true
}

provider "sftpgo" {
  host     = "http://localhost:8080"
  username = "admin"
  password = "password"

  headers = [
    {
      name  = "Proxy-Authorization"
      value = "Bearer ${data.google_service_account_id_token.sftp_iap_oidc.id_token}"
    }
  ]
}
