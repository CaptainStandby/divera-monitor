data "google_project" "project" {}

locals {
  region       = "europe-west3"
  project_name = trimprefix(data.google_project.project.id, "projects/")
}
