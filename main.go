package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/synadia-io/orbit.go/pcgroups"
)

const (
	STREAM      = "iot"
	CGROUP_NAME = "poc"
)

var memberPrefix = "member"

// Self explanatory
// Strong suggestion to use pod-name as the member name
// that way members are gaurenteed to be distinctly named
// and tracked easily.
// Neat little feature(Priority Groups added in NATS 2.11): Members of the same name form a priority group where only
// one pod will consume and the other will remain in standby for HA
func generateMemberName(currentMembersList []string) string {
	return fmt.Sprintf("%s-%02d", memberPrefix, len(currentMembersList))
}

// Simple set add implementation to display a list of imeis
// on the current member partition
type Set struct {
	set []string
}

func (s *Set) Add(str string) {
	if slices.Contains(s.set, str) {
		return
	}
	s.set = append(s.set, str)
	slices.Sort(s.set)
}

// Payload to get imei from the json data
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

func main() {
	// Conn setup
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

	// Used internally by the library to create consumers
	// AckWait and MaxAckPending allows member to endure
	// some downtime when other members are added/removed
	// and shuffling happens
	templateConsumerConfig := jetstream.ConsumerConfig{
		MaxAckPending: 1,
		AckWait:       1 * time.Second,
		AckPolicy:     jetstream.AckExplicitPolicy, // Requirement for elastic consumer
	}

	currentMembers, err := pcgroups.ListElasticActiveMembers(context.Background(), js, STREAM, CGROUP_NAME)
	if err != nil {
		log.Fatal(err)
	}

	memberName := generateMemberName(currentMembers)
	defer func() {
		// Remove self from group upon exit
		// as to not clog the members list
		// e.g. pods names beign used as members
		pcgroups.DeleteMembers(context.Background(), js, STREAM, CGROUP_NAME, []string{memberName})
	}()

	// This initiates the consumer, starts a goroutine
	var p Payload
	var imei string
	var listOfImeis Set
	consumerContext, err := pcgroups.ElasticConsume(
		context.Background(),
		js,
		STREAM,
		CGROUP_NAME,
		memberName,
		func(msg jetstream.Msg) {
			err := json.Unmarshal(msg.Data(), &p)
			if err != nil {
				log.Fatal("Failed to unmarshall payload")
			}
			imei = p.Data.ID
			listOfImeis.Add(imei)
			fmt.Printf("current_subject=%v, alloted=%s\n", msg.Subject(), strings.Join(listOfImeis.set, ", "))
			msg.Ack()
		},
		templateConsumerConfig,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer consumerContext.Stop()

	// This is where the actual consumption starts
	members, err := pcgroups.AddMembers(context.Background(), js, STREAM, CGROUP_NAME, []string{memberName})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(members)

	fmt.Println(<-consumerContext.Done())
}
