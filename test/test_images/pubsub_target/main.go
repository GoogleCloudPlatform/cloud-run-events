package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/knative-gcp/test/e2e/lib"
	"github.com/kelseyhightower/envconfig"
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

	fmt.Printf("Type to match: %q.\n", r.Type)
	fmt.Printf("Source to match: %q.\n", r.Source)

	// Create a timer
	duration, _ := strconv.Atoi(r.Time)
	timer := time.NewTimer(time.Second * time.Duration(duration))
	defer timer.Stop()
	go func() {
		<-timer.C
		// Write the termination message if time out
		fmt.Printf("time out to wait for event with source %q.\n", r.Source)
		if err := r.writeTerminationMessage(map[string]interface{}{
			"success": false,
		}); err != nil {
			fmt.Printf("failed to write termination message, %s.\n", err.Error())
		}
		os.Exit(0)
	}()

	if err := client.StartReceiver(context.Background(), r.Receive); err != nil {
		log.Fatal(err)
	}
}

type Receiver struct {
	Type string `envconfig:"TYPE" required:"true"`
	Source string `envconfig:"SOURCE" required:"true"`
	Time    string `envconfig:"TIME" required:"true"`
}

func (r *Receiver) Receive(event cloudevents.Event) {
	// Print out event received to log
	fmt.Printf("pubsub target received event\n")
	fmt.Printf("context of event is: %v\n", event.Context.String())

	incorrectAttributes := make(map[string]lib.PropPair)

	// Check type
	if event.Type() != r.Type {
		incorrectAttributes[r.Type] = lib.PropPair{Expected: r.Type, Received: event.Type()}
	}

	// Check source
	if event.Source() != r.Source {
		incorrectAttributes[lib.EventSource] = lib.PropPair{Expected: r.Source, Received: event.Source()}
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
			fmt.Println(k, "expected:", v.Expected, "got:", v.Received)
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
