job "delete-me" {
  region      = "FILL_IN"
  type        = "service"
  datacenters = ["FILL_IN"]

  group "svc" {

    network {
      mode = "bridge"

      port "http" {
        to = 6789
      }
    }

    service {

      port = "http"

    }

    task "server" {
      driver = "docker"

      config {
        args  = ["-text", "<head><meta http-equiv='Refresh' content='0; URL=https://eadderley.ca'></head>"]
        image = "hashicorp/http-echo:latest"
        ports = ["http"]
      }

      resources {}
    }
  }
}
