resource "google_pubsub_schema" "divera_alarm" {
  name       = "divera-alarm"
  type       = "PROTOCOL_BUFFER"
  definition = file("${path.module}/../proto/divera-alarm.proto")
}

resource "google_pubsub_topic" "divera_alarm" {
  name = "divera-alarm"

  schema_settings {
    schema   = google_pubsub_schema.divera_alarm.id
    encoding = "BINARY"
  }
}

resource "google_pubsub_subscription" "divera_alarm" {
  name  = "divera-alarm"
  topic = google_pubsub_topic.divera_alarm.id

  message_retention_duration = "1800s"
  retain_acked_messages      = true

  ack_deadline_seconds = 30

  expiration_policy {
    ttl = ""
  }
  retry_policy {
    minimum_backoff = "5s"
    maximum_backoff = "60s"
  }

  enable_message_ordering = true
}

resource "google_service_account" "subscriber" {
  account_id   = "alarm-subscriber-sa"
  display_name = "Divera Alarm Subscriber Service Account"
}

resource "google_service_account_key" "subscriber-key" {
  // TODO: This should not be done via Terraform, because it stores the private key in the state unencrypted.
  service_account_id = google_service_account.subscriber.id
  private_key_type   = "TYPE_GOOGLE_CREDENTIALS_FILE"
}

resource "google_pubsub_subscription_iam_binding" "pubsub_subscriber" {
  subscription = google_pubsub_subscription.divera_alarm.id
  role         = "roles/pubsub.subscriber"
  members = [
    google_service_account.subscriber.member
  ]
}
