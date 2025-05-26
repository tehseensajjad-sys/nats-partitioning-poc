package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/synadia-io/orbit.go/pcgroups"
)

const (
	STREAM      = "iot"
	CGROUP_NAME = "poc"
)

var memberName string = os.Getenv("HOSTNAME")

type Payload struct {
	Data struct {
		ID  string `json:"id"`
		GPS struct {
			Alt     int     `json:"alt"`
			Lat     float64 `json:"lat"`
			Lon     float64 `json:"lon"`
			Sat     int     `json:"sat"`
			Heading int     `json:"heading"`
		} `json:"gps"`
		From   int `json:"from"`
		Sensor struct {
			Ign         bool    `json:"ign"`
			RPM         int     `json:"rpm"`
			Itag        int     `json:"Itag"`
			Btry        float64 `json:"btry"`
			Dist        int     `json:"dist"`
			Speed       int     `json:"speed"`
			EcoDri      int     `json:"ecoDri"`
			Idling      bool    `json:"idling"`
			EngTemp     float64 `json:"engTemp"`
			Humidity    float64 `json:"humidity"`
			PedalPos    int     `json:"pedalPos"`
			CabinTemp   float64 `json:"cabinTemp"`
			FuelLevel   float64 `json:"fuelLevel"`
			FuelConsum  float64 `json:"fuelConsum"`
			FuelPercent float64 `json:"fuelPercent"`
		} `json:"sensor"`
		Network struct {
			GSM int `json:"gsm"`
		} `json:"network"`
		Timestamp int64 `json:"timestamp"`
	} `json:"data"`
	UUID      string `json:"uuid"`
	Timestamp int64  `json:"timestamp"`
}

func eventHandler(msg jetstream.Msg) {
	var p Payload
	err := json.Unmarshal(msg.Data(), &p)
	if err != nil {
		log.Fatal("Failed to unmarshall payload")
	}
	fmt.Printf("subject=%s imei=%s\n", msg.Subject(), p.Data.ID)
	msg.Ack()
}

func main() {
	opts, err := nats.NkeyOptionFromSeed("/Users/ma.afridi/.nats/dev-app.nk") // path to your .nk file
	if err != nil {
		log.Fatalf("Error reading NKEY file: %v", err)
	}

	nc, err := nats.Connect("nats://dev-nats.dev-dataplane.svc:4222", opts)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("Error getting JetStream context: %v", err)
	}

	templateConsumerConfig := jetstream.ConsumerConfig{
		MaxAckPending: 1,
		AckWait:       1 * time.Second,
		AckPolicy:     jetstream.AckExplicitPolicy,
	}

	fmt.Println("Running consumer")
	consumerContext, err := pcgroups.ElasticConsume(
		context.Background(),
		js,
		STREAM,
		CGROUP_NAME,
		memberName,
		eventHandler,
		templateConsumerConfig,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer consumerContext.Stop()
	members, err := pcgroups.AddMembers(context.Background(), js, STREAM, CGROUP_NAME, []string{memberName})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(members)

	isMember, active, err := pcgroups.ElasticIsInMembershipAndActive(context.Background(), js, STREAM, CGROUP_NAME, memberName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("is member=%v, is-active=%v\n", isMember, active)

	fmt.Println(<-consumerContext.Done())
}
