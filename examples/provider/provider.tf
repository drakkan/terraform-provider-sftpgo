# Configuration-based authentication
provider "sftpgo" {
  host     = "http://localhost:8080"
  username = "admin"
  password = "password"
}

# API key-based authentication with TLS verification disabled
provider "sftpgo" {
  host              = "https://sftpgo.example.com"
  api_key          = "your-api-key"
  tls_verification = false
}