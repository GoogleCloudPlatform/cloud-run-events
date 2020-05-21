package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/knative-gcp/test/e2e/lib"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"
)

const (
	eventSubject = "subject"
	eventData    = "data"
	eventType    = "type"
)

func main() {
	client, err := cloudevents.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	r := Receiver{}
	if err := envconfig.Process("", &r); err != nil {
		panic(err)
	}

	fmt.Printf("Waiting to receive event (timeout in %s seconds)...", r.Time)

	duration, _ := strconv.Atoi(r.Time)
	timer := time.NewTimer(time.Second * time.Duration(duration))
	defer timer.Stop()

	go func() {
		<-timer.C
		// Write the termination message if time out occurred
		fmt.Println("Timed out waiting for event from scheduler")
		if err := r.writeTerminationMessage(map[string]interface{}{
			"success": false,
		}); err != nil {
			fmt.Println("Failed to write termination message, got error:", err.Error())
		}
		os.Exit(0)
	}()

	if err := client.StartReceiver(context.Background(), r.Receive); err != nil {
		log.Fatal(err)
	}
}

type Receiver struct {
	Time          string `envconfig:"TIME" required:"true"`
	SubjectPrefix string `envconfig:"SUBJECT_PREFIX" required:"true"`
	Data          string `envconfig:"DATA" required:"true"`
	Type     string `envconfig:"TYPE" required:"true"`
}

type propPair struct {
	expected string
	received string
}

func (r *Receiver) Receive(event cloudevents.Event) {
	// Print out event received to log
	fmt.Printf("scheduler target received event\n")
	fmt.Printf(event.Context.String())

	incorrectAttributes := make(map[string]lib.PropPair)

	// Check subject prefix
	subject := event.Subject()
	if !strings.HasPrefix(subject, r.SubjectPrefix) {
		incorrectAttributes[lib.EventSubjectPrefix] = lib.PropPair{r.SubjectPrefix, subject}
	}

	// Check type
	evType := event.Type()
	if evType != r.Type {
		incorrectAttributes[lib.EventType] = lib.PropPair{r.Type, evType}
	}

	// Check data
	data := string(event.Data.([]uint8))
	if data != r.Data {
		incorrectAttributes[eventData] = lib.PropPair{r.Data, data}
	}

	if len(incorrectAttributes) == 0 {
		// Write the termination message.
		if err := r.writeTerminationMessage(map[string]interface{}{
			"success": true,
		}); err != nil {
			fmt.Printf("failed to write termination message, %s.\n", err)
		}
	} else {
		if err := r.writeTerminationMessage(map[string]interface{}{
			"success": false,
		}); err != nil {
			fmt.Printf("failed to write termination message, %s.\n", err)
		}
		for k, v := range incorrectAttributes {
			if k == lib.EventSubjectPrefix {
				fmt.Println(v.Received, "did not have expected prefix", v.Expected)
			} else {
				fmt.Println(k, "expected:", v.Expected, "got:", v.Received)
			}
		}
	}
	os.Exit(0)
}

func (r *Receiver) writeTerminationMessage(result interface{}) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("/dev/termination-log", b, 0644)
}
