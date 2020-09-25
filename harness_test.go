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
		true,   // auto-ack
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

func TestAMQP(t *testing.T) {
	protocol, publishAmqpServer, publishAmqpServerPort, publishUsername, publishPassword, publishURISuffix, subscribeAmqpServer, subscribeAmqpServerPort, subscribeURISuffix, subscribeUsername, subscribePassword, publishQ, subscribeQ, timeout := readEnvVars()

	publishAmqpURI := fmt.Sprintf("%s://%s:%s@%s:%s/%s", protocol, publishUsername, publishPassword, publishAmqpServer, publishAmqpServerPort, publishURISuffix)
	///////
	var testCases = []PactDetail{
		{"TestA", "hello", "hello"},
	}
	sendPayload := testCases[0].reqBody
	expectRecvPayload := testCases[0].respBody
	///////
	t.Logf("Publishing to %s\n", publishAmqpURI)
	conn, err := amqp.Dial(publishAmqpURI)
	if err != nil {
		t.Fatalf("Failed to connect to RabbitMQ: %v\n", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("Failed to open a channel: %v\n", err)
	}
	defer ch.Close()

	_, err = SendMsg(ch, publishQ, sendPayload)
	if err != nil {
		t.Fatalf("Failed to send message '%s': %v\n", sendPayload, err)
	}

	subscribeAmqpURI := fmt.Sprintf("amqps://%s:%s@%s:%s/%s", subscribeUsername, subscribePassword, subscribeAmqpServer, subscribeAmqpServerPort, subscribeURISuffix)
	if subscribeAmqpURI != publishAmqpURI {
		// we're reading from a different queue server; close the existing connection and open a new one
		ch.Close()
		conn.Close()
		conn, err := amqp.Dial(subscribeAmqpURI)
		if err != nil {
			t.Fatalf("Failed to connect to RabbitMQ: %v\n", err)
		}
		defer conn.Close()

		ch, err := conn.Channel()
		if err != nil {
			t.Fatalf("Failed to open a channel: %v\n", err)
		}
		defer ch.Close()
	}
	defer conn.Close()
	defer ch.Close()

	t.Logf("Subscribing to %s\n", subscribeAmqpURI)
	responsePayload, err := RecvMsg(ch, subscribeQ, timeout)
	if expectRecvPayload != responsePayload {
		t.Logf("Expected response '%s' doesn't match actual response '%s'\n", expectRecvPayload, responsePayload)
		t.Fail()
	} else {
		t.Log("Expected payload matches actual payload")
	}
}
