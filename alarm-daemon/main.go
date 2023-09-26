package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
	messages "github.com/CaptainStandby/divera-monitor/proto"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func toTime(t *messages.Alarm_Timestamp) time.Time {
	return time.Unix(t.Seconds, 0)
}

type trigger func(context.Context) error

const DEFAULT_LINGER_TIME = 15 * time.Minute
const DEFAULT_COMMAND_TIMEOUT = 30 * time.Second

type alarmTimer struct {
	lingerTime time.Duration
	lastUpdate time.Time
	storeTime  func(time.Time)
}

func (a *alarmTimer) update(t time.Time) {
	if t.After(a.lastUpdate) {
		a.lastUpdate = t
		if a.storeTime != nil {
			a.storeTime(t)
		}
	}
}

func (a *alarmTimer) standbyTime() time.Time {
	return a.lastUpdate.Add(a.lingerTime)
}

func (a *alarmTimer) isExpired() bool {
	return time.Now().After(a.standbyTime())
}

func (a *alarmTimer) isActive() bool {
	return !a.isExpired()
}

func (a *alarmTimer) standby() <-chan time.Time {
	if a.isExpired() {
		return make(<-chan time.Time)
	}
	return time.After(time.Until(a.standbyTime()))
}

func watcher(
	ctx context.Context,
	pipeline <-chan *messages.Alarm,
	switchOn, switchOff trigger,
	timer alarmTimer) {

	activate := func() {
		if timer.isActive() {
			log.Printf("watcher: alarm is active, switching on\n")
			if err := switchOn(ctx); err != nil {
				log.Printf("watcher: switchOn err: %v\n", err)
			}
		}
	}

	activate()

	for {
		select {
		case <-ctx.Done():
			log.Println("watcher: context done")
			return

		case msg := <-pipeline:
			timer.update(toTime(msg.Updated))
			activate()

		case <-timer.standby():
			log.Println("watcher: alarm has expired, switching off")
			if err := switchOff(ctx); err != nil {
				log.Printf("watcher: switchOff err: %v\n", err)
			}
		}
	}
}

func act(ctx context.Context, msg *messages.Alarm, pipeline chan<- *messages.Alarm) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case pipeline <- msg:
	}

	return nil
}

func handler(
	ctx context.Context,
	msg *pubsub.Message,
	act func(ctx context.Context, msg *messages.Alarm) error) {
	message := &messages.Alarm{}

	encoding := msg.Attributes["googclient_schemaencoding"]

	if encoding == "BINARY" {
		if err := proto.Unmarshal(msg.Data, message); err != nil {
			log.Printf("proto.Unmarshal err: %v\n", err)
			msg.Nack()
			return
		}
	} else if encoding == "JSON" {
		if err := protojson.Unmarshal(msg.Data, message); err != nil {
			log.Printf("proto.Unmarshal err: %v\n", err)
			msg.Nack()
			return
		}
	} else {
		log.Printf("Unknown message type(%s), nacking\n", encoding)
		msg.Nack()
		return
	}

	if err := act(ctx, message); err != nil {
		log.Printf("could not process message, nacking: %v\n", err)
		msg.Nack()
		return
	}

	msg.Ack()
}

func startListening(ctx context.Context, sub *pubsub.Subscription, watcher func(ctx context.Context, pipeline <-chan *messages.Alarm)) {
	pipeline := make(chan *messages.Alarm, 10)

	go watcher(ctx, pipeline)

	go func() {
		defer close(pipeline)
		log.Println("Start receiving messages")

		err := sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
			handler(ctx, m, func(ctx context.Context, msg *messages.Alarm) error {
				return act(ctx, msg, pipeline)
			})
		})
		if err != nil {
			log.Fatalf("sub.Receive: %s", err)
		}
	}()
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
	lingerTime := DEFAULT_LINGER_TIME
	if val, ok := os.LookupEnv("LINGER_TIME"); ok {
		v, err := time.ParseDuration(val)
		lingerTime = v
		if err != nil {
			log.Fatalf("LINGER_TIME environment variable is not a valid duration: %v", err)
		}
	}
	switchOnCmd := os.Getenv("SWITCH_ON_CMD")
	if switchOnCmd == "" {
		log.Fatal("SWITCH_ON_CMD environment variable is not set")
	}
	switchOffCmd := os.Getenv("SWITCH_OFF_CMD")
	if switchOffCmd == "" {
		log.Fatal("SWITCH_OFF_CMD environment variable is not set")
	}
	commandTimeout := DEFAULT_COMMAND_TIMEOUT
	if val, ok := os.LookupEnv("COMMAND_TIMEOUT"); ok {
		v, err := time.ParseDuration(val)
		commandTimeout = v
		if err != nil {
			log.Fatalf("COMMAND_TIMEOUT environment variable is not a valid duration: %v", err)
		}
	}
	lastAlarmFile := os.Getenv("LAST_ALARM_FILE")

	ctx, cancel := context.WithCancel(context.Background())

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

	timer := alarmTimer{
		lingerTime: lingerTime,
		lastUpdate: loadLastAlarmTime(lastAlarmFile),
		storeTime:  func(t time.Time) { storeLastAlarmTime(lastAlarmFile, t) },
	}

	startListening(ctx, sub, func(ctx context.Context, pipeline <-chan *messages.Alarm) {
		watcher(ctx, pipeline, func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, commandTimeout)
			defer cancel()

			log.Println("switching on")
			return executeCommand(ctx, switchOnCmd)
		}, func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, commandTimeout)
			defer cancel()

			log.Println("switching off")
			return executeCommand(ctx, switchOffCmd)
		}, timer)
	})

	waitForShutdown(client, cancel)
}

func storeLastAlarmTime(lastAlarmFile string, t time.Time) {
	if lastAlarmFile != "" {
		if f, err := os.OpenFile(lastAlarmFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err == nil {
			defer f.Close()
			if _, err := fmt.Fprintf(f, "%s\n", t.UTC().Format(time.RFC3339)); err != nil {
				log.Printf("could not write last alarm time to %s: %v\n", lastAlarmFile, err)
			}
		} else {
			log.Printf("could not open last alarm file %s: %v\n", lastAlarmFile, err)
		}
	}
}

func loadLastAlarmTime(lastAlarmFile string) time.Time {
	if lastAlarmFile == "" {
		return time.Unix(0, 0)
	}

	if f, err := os.Open(lastAlarmFile); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		if scanner.Scan() {
			if t, err := time.Parse(time.RFC3339, scanner.Text()); err == nil {
				log.Printf("last alarm time: %s\n", t)
				return t
			} else {
				log.Printf("could not parse last alarm time from %s: %v\n", lastAlarmFile, err)
			}
		} else {
			log.Printf("could not read last alarm time from %s: %v\n", lastAlarmFile, err)
		}
	} else {
		log.Printf("could not open last alarm file %s: %v\n", lastAlarmFile, err)
	}

	return time.Unix(0, 0)
}

func executeCommand(ctx context.Context, command string) error {
	cmd := exec.CommandContext(ctx, command)
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("cmd.StdoutPipe: %v\n", err)
	}
	go func() {
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			log.Printf("[%s] %s", command, scanner.Text())
		}
	}()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cmd.Start(%s): %w", command, err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("cmd.Wait(%s): %w", command, err)
	}

	return nil
}

func waitForShutdown(client *pubsub.Client, cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	<-c
	log.Println("Shutdown requested")

	cancel()
	err := client.Close()
	if err != nil {
		log.Fatalf("client.Close: %s", err)
	}

	log.Println("Shutdown complete")
	os.Exit(0)
}
