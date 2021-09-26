resource "docker_image" "blocksvc-image" {
  name = "blocksvc"

  keep_locally = false
  force_remove = true

  build {
    path = "./.."
    dockerfile = "Containerfile"
    tag = [
      "blocksvc:v1"
    ]
    // NOTE: remove intermediate containers
    force_remove = true
  }
}


resource "docker_container" "blocksvc-container" {
  name = "blocksvc-instance"
  image = "blocksvc${ docker_image.blocksvc-image.repo_digest }"
}
