
locals {
  required_services = [
    "cloudfunctions.googleapis.com",
    "run.googleapis.com",
    "artifactregistry.googleapis.com",
    "cloudbuild.googleapis.com"
  ]
}

resource "google_project_service" "services" {
  for_each = toset(local.required_services)
  service  = each.value
}

resource "google_service_account" "publisher" {
  account_id   = "gcf-sa"
  display_name = "Cloud Function Service Account"
}

resource "google_pubsub_topic_iam_binding" "pubsub_publisher" {
  topic = google_pubsub_topic.divera_alarm.id
  role  = "roles/pubsub.publisher"
  members = [
    google_service_account.publisher.member
  ]
}

resource "google_cloudfunctions2_function" "alarm_ingress" {
  depends_on = [google_project_service.services]
  name       = "alarm-ingress"
  location   = local.region

  build_config {
    runtime     = "go120"
    entry_point = "HandleAlarm"
    source {
      storage_source {
        bucket = google_storage_bucket.bucket.name
        object = google_storage_bucket_object.alarm_ingress_source.name
      }
    }
  }

  service_config {
    max_instance_count               = 1
    min_instance_count               = 0
    available_memory                 = "256M"
    timeout_seconds                  = 15
    max_instance_request_concurrency = 16
    available_cpu                    = "1"
    ingress_settings                 = "ALLOW_ALL"
    all_traffic_on_latest_revision   = true
    service_account_email            = google_service_account.publisher.email
    environment_variables = {
      PROJECT_ID = data.google_project.project.project_id
      TOPIC_NAME = google_pubsub_topic.divera_alarm.name
    }
  }
}

resource "google_cloud_run_service_iam_binding" "public_invoker" {
  project  = google_cloudfunctions2_function.alarm_ingress.project
  location = google_cloudfunctions2_function.alarm_ingress.location
  service  = google_cloudfunctions2_function.alarm_ingress.name
  role     = "roles/run.invoker"
  members  = ["allUsers"]
}
