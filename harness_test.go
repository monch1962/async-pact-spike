package main

import (
	"fmt"

	"github.com/streadway/amqp"
	"log"
	"os"
	"testing"
)

func TestFailing(t *testing.T) {
	t.Fail()
}

func readEnvVars() (string, string, string, string, string, string, string, string) {
	publishAmqpServer := os.Getenv("PUBLISH_AMQP_SERVER")
	publishAmqpTCPPort := os.Getenv("PUBLISH_AMQP_SERVER_TCP")
	if publishAmqpTCPPort == "" {
		publishAmqpTCPPort = "5672"
	}
	subscribeAmqpServer := os.Getenv("SUBSCRIbE_AMQP_SERVER")
	if subscribeAmqpServer == "" {
		subscribeAmqpServer = publishAmqpServer
	}
	subscribeAmqpTCPPort := os.Getenv("SUBSCRIBE_AMQP_SERVER_TCP")
	if subscribeAmqpTCPPort == "" {
		subscribeAmqpTCPPort = publishAmqpTCPPort
	}
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	publishQ := os.Getenv("PUBLISH_Q")
	subscribeQ := os.Getenv("SUBSCRIBE_Q")
	return publishAmqpServer,publishAmqpTCPPort,subscribeAmqpServer, subscribeAmqpTCPPort,username,password,publishQ,subscribeQ
}

func SendMsg(ch *amqp.Channel, queue string, payload string) (string, error) {
	publishAmqpServer, publishAmqpServerPort, subscribeAmqpServer, subscribeAmqpServerPort, username, password, publishQ, subscribeQ := readEnvVars()
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

func RecvMsg(ch *amqp.Channel, queue string) (string, error) {
	_, _, subscribeAmqpServer, subscribeAmqpServerPort, username, password, _, subscribeQ := readEnvVars()
	amqpURI := fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password,subscribeAmqpServer, subscribeAmqpServerPort)
	//conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	ch, err = conn.Channel()
	if err != nil {
		return "", fmt.Errorf("Failed to open a channel")
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		subscribeQ, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return "", fmt.Errorf( "Failed to declare a queue")
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

	forever := make(chan bool)
	var msg string

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			msg = string(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	return msg, nil
}

func TestAMQP(t *testing.T) {
	sendQueue := "Hello"
	sendPayload := "Hello"
	recvQueue := "Hello"
	expectRecvPayload := "Hello"
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
	responsePayload, err := RecvMsg(ch, recvQueue)
	if responsePayload != expectRecvPayload {
		t.Logf("Expected response '%s' doesn't match actual response '%s'\n", expectRecvPayload, responsePayload)
		t.Fail()
	}
}
