package main

import (
	"fmt"

	"github.com/streadway/amqp"
	//"log"
	"os"
	"testing"
)

func TestFailing(t *testing.T) {
	t.Fail()
}

func readEnvVars() (string, string, string, string, string, string) {
	amqpServer := os.Getenv("AMQP_SERVER")
	amqpTCPPort := os.Getenv("AMQP_SERVER_TCP")
	if amqpTCPPort == "" {
		amqpTCPPort = "5672"
	}
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	publishQ := os.Getenv("PUBLISH_Q")
	subscribeQ := os.Getenv("SUBSCRIBE_Q")
	return amqpServer,amqpTCPPort,username,password,publishQ,subscribeQ
}

func SendMsg(ch *amqp.Channel, queue string, payload string) (string, error) {
	q, err := ch.QueueDeclare(
		queue, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return "",fmt.Errorf("Failed to declare a queue:%v", err)
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
		return "",fmt.Errorf("Failed to publish a message: %v", err)
	}
	return body, nil
}

func TestAMQP(t *testing.T) {
	sendQueue := "Hello"
	sendPayload := "Hello"
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Fatalf("Failed to connect to RabbitMQ: %v\n", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("Failed to open a channel: %v\n", err)
	}

	defer ch.Close()

	_, err = SendMsg(ch, sendQueue, sendPayload)
	if err != nil {
		t.Fatalf("Failed to send message '%s': %v\n", sendPayload, err)
	}
}
