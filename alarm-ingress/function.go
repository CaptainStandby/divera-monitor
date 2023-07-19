package function

import (
	"context"
	"log"
	"os"

	"github.com/CaptainStandby/divera-monitor/alarm-ingress/alarm"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

	"cloud.google.com/go/pubsub"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

func init() {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		projectID = pubsub.DetectProjectID
	}
	topicName := os.Getenv("TOPIC_NAME")
	if topicName == "" {
		log.Fatal("TOPIC_NAME environment variable is not set")
	}

	ctx := context.Background()

	cred, err := google.FindDefaultCredentials(ctx, pubsub.ScopePubSub)
	if err != nil {
		log.Fatalf("google.FindDefaultCredentials: %s", err)
	}

	client, err := pubsub.NewClient(ctx, projectID, option.WithCredentials(cred))
	if err != nil {
		log.Fatalf("pubsub.NewClient: %s", err)
	}

	topic := client.Topic(topicName)
	if topic == nil {
		log.Fatalf("client.Topic(%s) returned nil", topicName)
	}

	functions.HTTP("HandleAlarm", alarm.BuildHandler(topic.Publish))
}
