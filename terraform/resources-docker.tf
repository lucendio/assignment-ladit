resource "docker_image" "blocksvc-image" {
  name = "blocksvc:v1"

  keep_locally = false
  force_remove = true

  build {
    path = "./.."
    dockerfile = "Containerfile"
    // NOTE: remove intermediate containers
    force_remove = true
  }
}


resource "docker_container" "blocksvc-container" {
  name = "blocksvc-instance"
  image = docker_image.blocksvc-image.name

  ports {
    internal = 8080
    external = 8080
  }

  networks_advanced {
    name = docker_network.test-network.name
  }

  env= [
    "ENV_NAME=production",
    "PORT=8080",
    "HOST=0.0.0.0",
    // NOTE: obviously, this kind of information is not supposed to be
    //       anywhere near a version control system
    "ACCESS_TOKEN=sensitive"
  ]
}


resource "docker_network" "test-network" {
  name = "test"
  driver = "bridge"
}
