package function

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/pkg/errors"

	"cloud.google.com/go/pubsub"
	messages "github.com/CaptainStandby/divera-monitor/proto"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/proto"
)

type jsonAlarm struct {
	ID               int64   `json:"id"`
	ForeignID        string  `json:"foreign_id"`
	Title            string  `json:"title"`
	Text             string  `json:"text"`
	Address          string  `json:"address"`
	Lat              float64 `json:"lat"`
	Lng              float64 `json:"lng"`
	Priority         bool    `json:"priority"`
	NotificationType int     `json:"notification_type"`
	Created          int64   `json:"ts_create"`
	Updated          int64   `json:"ts_update"`
}

func convertToProto(alarm *jsonAlarm) *messages.Alarm {
	return &messages.Alarm{
		Id:        alarm.ID,
		ForeignId: alarm.ForeignID,
		Title:     alarm.Title,
		Text:      alarm.Text,
		Address:   alarm.Address,
		Position:  &messages.Alarm_LatLng{Latitude: alarm.Lat, Longitude: alarm.Lng},
		Priority:  alarm.Priority,
		Created:   &messages.Alarm_Timestamp{Seconds: alarm.Created},
		Updated:   &messages.Alarm_Timestamp{Seconds: alarm.Updated},
	}
}

var client *pubsub.Client
var topic *pubsub.Topic

func init() {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Fatal("PROJECT_ID environment variable is not set")
	}
	topicName := os.Getenv("TOPIC_NAME")
	if topicName == "" {
		log.Fatal("TOPIC_NAME environment variable is not set")
	}

	ctx := context.Background()

	cred, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		log.Fatalf("google.FindDefaultCredentials: %s", err)
	}

	client, err = pubsub.NewClient(ctx, projectID, option.WithCredentials(cred))
	if err != nil {
		log.Fatalf("pubsub.NewClient: %s", err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	go func() {
		defer client.Close()
		<-c
	}()

	topic = client.Topic(topicName)
	if topic == nil {
		log.Fatalf("client.Topic(%s) returned nil", topicName)
	}

	functions.HTTP("HandleAlarm", handleAlarm)
}

func pushAlarm(ctx context.Context, alarm *jsonAlarm) error {

	data, err := proto.Marshal(convertToProto(alarm))
	if err != nil {
		return errors.Wrap(err, "proto.Marshal() failed")
	}

	res := topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})
	_, err = res.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "Publish() failed")
	}

	return nil
}

func handleAlarm(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	msg := &jsonAlarm{}
	err := decoder.Decode(msg)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	err = pushAlarm(ctx, msg)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
