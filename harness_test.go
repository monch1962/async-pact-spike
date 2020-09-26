package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"log"
	"os"
	"testing"

	"github.com/streadway/amqp"
	"github.com/remeh/sizedwaitgroup"
)

/*func TestFailing(t *testing.T) {
	t.Fail()
}*/

func readEnvVars() (string, string, string, string, string, string, string, string, string, string, string, string, string, int64) {
	timeout, err := strconv.ParseInt(os.Getenv("TIMEOUT"), 10, 0)
	if err != nil {
		timeout = 10000
	}
	protocol := strings.ToLower(os.Getenv("PROTOCOL"))
	if protocol != "amqp" && protocol != "amqps" {
		protocol = "amqps"
	}

	publishAmqpServer := os.Getenv("PUBLISH_AMQP_SERVER")
	publishAmqpTCPPort := os.Getenv("PUBLISH_AMQP_SERVER_TCP")
	if publishAmqpTCPPort == "" {
		if protocol == "amqp" {
			publishAmqpTCPPort = "5672"
		} else {
			publishAmqpTCPPort = "5671"
		}
	}
	publishURISuffix := os.Getenv("PUBLISH_URI_SUFFIX")
	subscribeAmqpServer := os.Getenv("SUBSCRIBE_AMQP_SERVER")
	if subscribeAmqpServer == "" {
		subscribeAmqpServer = publishAmqpServer
	}
	subscribeAmqpTCPPort := os.Getenv("SUBSCRIBE_AMQP_SERVER_TCP")
	if subscribeAmqpTCPPort == "" {
		subscribeAmqpTCPPort = publishAmqpTCPPort
	}
	subscribeURISuffix := os.Getenv("SUBSCRIBE_URI_SUFFIX")
	if subscribeURISuffix == "" {
		subscribeURISuffix = publishURISuffix
	}
	publishUsername := os.Getenv("PUBLISH_USERNAME")
	if publishUsername == "" {
		publishUsername = "guest"
	}
	publishPassword := os.Getenv("PUBLISH_PASSWORD")
	if publishPassword == "" {
		publishPassword = "guest"
	}
	subscribeUsername := os.Getenv("SUBSCRIBE_USERNAME")
	if subscribeUsername == "" {
		subscribeUsername = publishUsername
	}
	subscribePassword := os.Getenv("SUBSCRIBE_PASSWORD")
	if subscribePassword == "" {
		subscribePassword = publishPassword
	}
	publishQ := os.Getenv("PUBLISH_Q")
	subscribeQ := os.Getenv("SUBSCRIBE_Q")
	return protocol, publishAmqpServer, publishAmqpTCPPort, publishUsername, publishPassword, publishURISuffix, subscribeAmqpServer, subscribeAmqpTCPPort, subscribeURISuffix, subscribeUsername, subscribePassword, publishQ, subscribeQ, timeout
}

func SendMsg(ch *amqp.Channel, queue string, payload string) (string, error) {
	log.Printf("Attempting to write payload '%s' to queue '%s'\n", payload, queue)
	q, err := ch.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return "", fmt.Errorf("Failed to declare a queue:%v", err)
	}

	body := payload
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		return "", fmt.Errorf("Failed to publish a message: %v", err)
	}
	return body, nil
}

func RecvMsg(ch *amqp.Channel, queue string, timeout int64) (string, error) {
	log.Printf("Reading from queue '%s'\n", queue)
	ch.Qos(1, 0, false)
	q, err := ch.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return "", fmt.Errorf("Failed to declare a queue")
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,   // auto-ack - set to false to ensure only a single message gets read off the queue at a time
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return "", fmt.Errorf("Failed to register a consumer")
	}
	//log.Printf("msgs: %v\n", msgs)
	var msg string

	duration := time.Duration(timeout) * time.Millisecond
	timer := time.NewTimer(duration)
	for {
		select {
		case d := <-msgs:
			//log.Printf("d: %v\n", d)
			timer.Reset(duration)
			log.Printf("Received a message: '%s'\n", d.Body)
			msg = string(d.Body)
		case <-timer.C:
			log.Println("Timeout !")
			return "", fmt.Errorf("Timeout waiting for message to appear")
		}

		return msg, nil
	}

}

type PactDetail struct {
	testName string
	reqBody  string
	respBody interface{}
}

func getAmqpChannel(uri string) (*amqp.Channel, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Failed to open a channel: %v", err)
	}
	return ch, nil
}

func TestAMQPConsumerProvider(t *testing.T) {
	protocol, publishQServer, publishQServerPort, publishUsername, publishPassword, publishURISuffix, subscribeQServer, subscribeQServerPort, subscribeURISuffix, subscribeUsername, subscribePassword, publishQ, subscribeQ, timeout := readEnvVars()

	publishURI := fmt.Sprintf("%s://%s:%s@%s:%s/%s", protocol, publishUsername, publishPassword, publishQServer, publishQServerPort, publishURISuffix)
	subscribeURI := fmt.Sprintf("%s://%s:%s@%s:%s/%s", protocol, subscribeUsername, subscribePassword, subscribeQServer, subscribeQServerPort, subscribeURISuffix)

	pubCh, err := getAmqpChannel(publishURI)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	var subCh *amqp.Channel
	if subscribeURI != publishURI {
		// we're reading from a different queue server; need a different channel
		subCh, err = getAmqpChannel(subscribeURI)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
	} else {
		// use the same channel for subscribe as for publish
		subCh = pubCh
	}
	t.Logf("Publishing to %s\n", publishURI)
	t.Logf("Subscribing to %s\n", subscribeURI)
	///////
	var testCases = []PactDetail{
		{"TestA", "hello", "hello"},
		{"TestB", "there", "there"},
		{"TestC", "everywhere", "everywhere"},
	}
	///////
	concurrency := 1
	swg := sizedwaitgroup.New(concurrency)

	for _, tc := range testCases {
		swg.Add()
		go func(t *testing.T, tc PactDetail) {
			defer swg.Done()
			t.Run(tc.testName, func(t *testing.T) {

				_, err := SendMsg(pubCh, publishQ, tc.reqBody)
				if err != nil {
					t.Logf("Failed to send message '%s': %v\n", tc.reqBody, err)
					t.Fail()
				}

				responsePayload, _ := RecvMsg(subCh, subscribeQ, timeout)
				if tc.respBody != responsePayload {
					t.Logf("Expected response '%s' doesn't match actual response '%s'\n", tc.respBody, responsePayload)
					t.Fail()
				} else {
					t.Log("Expected payload matches actual payload")
				}
			})
		}(t, tc)
	}
	swg.Wait()
}
