package alarm

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"cloud.google.com/go/pubsub"
	messages "github.com/CaptainStandby/divera-monitor/proto"
	"google.golang.org/protobuf/proto"
)

type jsonAlarm struct {
	ID               int64  `json:"id"`
	ForeignID        string `json:"foreign_id"`
	Title            string `json:"title"`
	Text             string `json:"text"`
	Address          string `json:"address"`
	Lat              string `json:"lat"`
	Lng              string `json:"lng"`
	Priority         int    `json:"priority"`
	NotificationType int    `json:"notification_type"`
	Created          int64  `json:"ts_create"`
	Updated          int64  `json:"ts_update"`
}

func convertToProto(alarm *jsonAlarm) *messages.Alarm {
	return &messages.Alarm{
		Id:        alarm.ID,
		ForeignId: alarm.ForeignID,
		Title:     alarm.Title,
		Text:      alarm.Text,
		Address:   alarm.Address,
		Position:  &messages.Alarm_LatLng{Latitude: /*alarm.Lat*/ 0, Longitude: /*alarm.Lng*/ 0},
		Priority:  alarm.Priority != 0,
		Created:   &messages.Alarm_Timestamp{Seconds: alarm.Created},
		Updated:   &messages.Alarm_Timestamp{Seconds: alarm.Updated},
	}
}

func pushAlarm(
	ctx context.Context,
	alarm *jsonAlarm,
	publish func(context.Context, *pubsub.Message) *pubsub.PublishResult) error {

	data, err := proto.Marshal(convertToProto(alarm))
	if err != nil {
		return errors.Wrap(err, "proto.Marshal() failed")
	}

	res := publish(ctx, &pubsub.Message{
		Data: data,
	})
	_, err = res.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "Publish() failed")
	}

	return nil
}

func handle(w http.ResponseWriter, r *http.Request, pushAlarm func(context.Context, *jsonAlarm) error) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("body: %s", body)

	msg := &jsonAlarm{}
	err = json.Unmarshal(body, msg)
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

func BuildHandler(publish func(context.Context, *pubsub.Message) *pubsub.PublishResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handle(w, r, func(ctx context.Context, alarm *jsonAlarm) error {
			return pushAlarm(ctx, alarm, publish)
		})
	}
}
