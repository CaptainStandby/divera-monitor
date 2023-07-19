output "public_url" {
  value = google_cloudfunctions2_function.alarm_ingress.url
}

output "subscription_name" {
  value = google_pubsub_subscription.divera_alarm.name
}

output "topic_name" {
  value = google_pubsub_topic.divera_alarm.name
}

output "subscriber_key_id" {
  value = google_service_account_key.subscriber_key.id
}

output "subscriber_email" {
  value = google_service_account.subscriber.email
}

output "subscriber_id" {
  value = google_service_account.subscriber.unique_id
}
