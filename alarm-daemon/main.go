package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/pubsub"
	messages "github.com/CaptainStandby/divera-monitor/proto"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func act(ctx context.Context, msg *messages.Alarm) {
	// TODO: implement
}

func handler(ctx context.Context, msg *pubsub.Message) {
	message := &messages.Alarm{}

	encoding := msg.Attributes["googclient_schemaencoding"]

	if encoding == "BINARY" {
		if err := proto.Unmarshal(msg.Data, message); err != nil {
			fmt.Printf("proto.Unmarshal err: %v\n", err)
			msg.Nack()
			return
		}
		fmt.Printf("Received a binary-encoded message:\n%#v\n", message)
	} else if encoding == "JSON" {
		if err := protojson.Unmarshal(msg.Data, message); err != nil {
			fmt.Printf("proto.Unmarshal err: %v\n", err)
			msg.Nack()
			return
		}
		fmt.Printf("Received a JSON-encoded message:\n%#v\n", message)
	} else {
		fmt.Printf("Unknown message type(%s), nacking\n", encoding)
		msg.Nack()
		return
	}

	act(ctx, message)
	msg.Ack()
}

func main() {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		projectID = pubsub.DetectProjectID
	}
	subscriptionName := os.Getenv("SUBSCRIPTION_NAME")
	if subscriptionName == "" {
		log.Fatal("SUBSCRIPTION_NAME environment variable is not set")
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
	defer client.Close()

	sub := client.Subscription(subscriptionName)
	if sub == nil {
		log.Fatalf("client.Subscription(%s) returned nil", subscriptionName)
	}

	go func() {
		log.Println("Start receiving messages")

		err = sub.Receive(ctx, handler)
		if err != nil {
			log.Fatalf("sub.Receive: %s", err)
		}
	}()

	waitForShutdown(client)
}

func waitForShutdown(client *pubsub.Client) {
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	<-c
	log.Println("Shutdown requested")

	err := client.Close()
	if err != nil {
		log.Fatalf("client.Close: %s", err)
	}

	log.Println("Shutdown complete")
	os.Exit(0)
}
