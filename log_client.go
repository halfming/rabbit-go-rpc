package main

import (
        "fmt"
        "log"
        "math/rand"
        "os"
//        "strconv"
        "strings"
        "time"

        "github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
        if err != nil {
                log.Fatalf("%s: %s", msg, err)
                panic(fmt.Sprintf("%s: %s", msg, err))
        }
}

func randomString(l int) string {
        bytes := make([]byte, l)
        for i := 0; i < l; i++ {
                bytes[i] = byte(randInt(65, 90))
        }
        return string(bytes)
}

func randInt(min int, max int) int {
        return min + rand.Intn(max-min)
}

func fibonacciRPC(n string, host string) (res string, err error) {
        conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
        failOnError(err, "Failed to connect to RabbitMQ")
        defer conn.Close()

        ch, err := conn.Channel()
        failOnError(err, "Failed to open a channel")
        defer ch.Close()

        q, err := ch.QueueDeclare(
                "",    // name
                false, // durable
                false, // delete when usused
                true,  // exclusive
                false, // noWait
                nil,   // arguments
        )
        failOnError(err, "Failed to declare a queue")

        msgs, err := ch.Consume(
                q.Name, // queue
                "",     // consumer
                true,   // auto-ack
                false,  // exclusive
                false,  // no-local
                false,  // no-wait
                nil,    // args
        )
        failOnError(err, "Failed to register a consumer")

        corrId := randomString(32)

        err = ch.Publish(
                "",          // exchange
                host, // routing key
                false,       // mandatory
                false,       // immediate
                amqp.Publishing{
                        ContentType:   "text/plain",
                        CorrelationId: corrId,
                        ReplyTo:       q.Name,
                        Body:          []byte(n),
                })
        failOnError(err, "Failed to publish a message")

        for d := range msgs {
                if corrId == d.CorrelationId {
                        res = string(d.Body)
                        //failOnError(err, "Failed to convert body to integer")
                        break
                }
        }

        return
}

func main() {
        rand.Seed(time.Now().UTC().UnixNano())

        n := bodyFrom(os.Args)
        if (n == "error") {
            os.Exit(1)
        }
        //log.Printf(" [x] Requesting fib(%d)", n)
		hosts := []string{"localhost.localdomain"}
		for i := range hosts {
			host := strings.Replace(hosts[i],".","_",-1)
			res, err := fibonacciRPC(n,host)
            failOnError(err, "Failed to handle RPC request")

            log.Printf(" show " + res)
		}
}

func bodyFrom(args []string) string {
        var s string
        if (len(args) < 2) || os.Args[1] == "" || os.Args[2] == ""{
                s = "error"
        } else {
                s = strings.Join(args[1:], " ")
        }
        // n, err := s
        // failOnError(err, "Failed to convert arg to integer")
        return s
}