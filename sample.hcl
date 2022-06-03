job "delete-me" {
  region      = "FILL_IN"
  type        = "service"
  datacenters = ["FILL_IN"]

  }

  group "svc" {

    network {
      mode = "bridge"

      port "http" {
        to = 6789
      }
    }


      port = "http"

      check {
        type     = "tcp"
        interval = "10s"
        timeout  = "5s"
      }
    }

    task "server" {
      driver = "docker"

      config {
        args  = ["-text", "<head><meta http-equiv='Refresh' content='0; URL=https://eadderley.ca'></head>"]
        image = "hashicorp/http-echo"
        ports = ["http"]
      }

      resources {}
    }
  }
}
