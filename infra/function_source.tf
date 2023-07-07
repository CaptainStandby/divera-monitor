
data "archive_file" "alarm_ingress_source" {
  type        = "zip"
  output_path = "${path.module}/.tmp/alarm-ingress.zip"
  source_dir  = "${path.module}/../alarm-ingress"
  excludes = [
    "cmd",
    "function_test.go"
  ]
}

resource "google_storage_bucket" "bucket" {
  name                        = "${local.project_name}-gcf-source"
  location                    = local.region
  force_destroy               = true
  uniform_bucket_level_access = true
  storage_class               = "STANDARD"
  public_access_prevention    = "enforced"

  versioning {
    enabled = false
  }
}

resource "google_storage_bucket_object" "alarm_ingress_source" {
  name   = "${data.archive_file.alarm_ingress_source.output_sha}.zip"
  bucket = google_storage_bucket.bucket.id
  source = data.archive_file.alarm_ingress_source.output_path
}
