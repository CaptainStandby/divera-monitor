output "public_url" {
  value = google_cloudfunctions2_function.alarm_ingress.url
}

output "subscription_name" {
  value = google_pubsub_subscription.divera_alarm.name
}

output "topic_name" {
  value = google_pubsub_topic.divera_alarm.name
}

output "subscriber_key" {
  value     = google_service_account_key.subscriber-key.private_key
  sensitive = true
}
