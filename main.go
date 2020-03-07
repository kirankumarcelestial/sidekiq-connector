// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Copyright (c) Keiran Smith 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jrallison/go-workers"
	"github.com/openfaas-incubator/connector-sdk/types"
)

type connectorConfig struct {
	upstreamTimeout time.Duration
	queues          []string
	printResponse   bool
	rebuildInterval time.Duration
	redis_host      string
}

func main() {
  log.Println("Entering 1")
	config := buildConnectorConfig()
  log.Println("Entering 2")
	//topicMap := types.NewTopicMap()
  log.Println("Entering 3")
	//lookupBuilder := types.FunctionLookupBuilder{
	//	Client: types.MakeClient(config.upstreamTimeout),
	//}
  log.Println("Entering 4")
	creds := types.GetCredentials()
	controllerconfig := &types.ControllerConfig{
		RebuildInterval:   time.Millisecond * 1000,
		GatewayURL:        "http://a56d6c9b55f2011eaae4402584498c9a-350195318.us-west-2.elb.amazonaws.com:8080",
		PrintResponse:     true,
		PrintResponseBody: true,
	}
  log.Println("Entering 5")
	controller := types.NewController(creds, controllerconfig)

	receiver := ResponseReceiver{}
	controller.Subscribe(&receiver)

	controller.BeginMapBuilder()
  log.Println("Entering 6")
	//ticker := time.NewTicker(config.rebuildInterval)
  //go synchronizeLookups(ticker, &lookupBuilder, &topicMap)
  log.Println("Entering 7")
	workers.Configure(map[string]string{
		"server":   config.redis_host,
		"database": "0",
		"pool":     "30",
		"process":  "1",
	})
  log.Println("Entering 8")
	for _, queue := range config.queues {
		handler := makeMessageHandler(controller, queue)
		workers.Process(queue, handler, 1)
	}
  log.Println("Entering 9")
	workers.Run()
}

func synchronizeLookups(ticker *time.Ticker,
	lookupBuilder *types.FunctionLookupBuilder,
	topicMap *types.TopicMap) {

	for {
		<-ticker.C
		lookups, err := lookupBuilder.Build()
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("Syncing topic map")
		topicMap.Sync(&lookups)
	}
}

func makeMessageHandler(controller *types.Controller, queue string) func(msg *workers.Msg) {

	mcb := func(msg *workers.Msg) {
		msgJson, err := json.Marshal(msg.Args)

		if err != nil {
			log.Fatal(err.Error())
		}
    log.Println("Entering %s", msgJson)
		controller.Invoke(queue, &msgJson)
	}
	return mcb
}

func buildConnectorConfig() connectorConfig {

	redis := "redis_host"
	if val, exists := os.LookupEnv("redis_host"); exists {
		redis = val
	}

	queues := []string{}
	if val, exists := os.LookupEnv("queues"); exists {
		for _, topic := range strings.Split(val, ",") {
			if len(topic) > 0 {
				queues = append(queues, topic)
			}
		}
	}
	if len(queues) == 0 {
		log.Fatal(`Provide a list of queues i.e. queues="payment_published,slack_joined"`)
	}

	upstreamTimeout := time.Second * 30
	rebuildInterval := time.Second * 3

	if val, exists := os.LookupEnv("upstream_timeout"); exists {
		parsedVal, err := time.ParseDuration(val)
		if err == nil {
			upstreamTimeout = parsedVal
		}
	}

	if val, exists := os.LookupEnv("rebuild_interval"); exists {
		parsedVal, err := time.ParseDuration(val)
		if err == nil {
			rebuildInterval = parsedVal
		}
	}

	printResponse := false
	if val, exists := os.LookupEnv("print_response"); exists {
		printResponse = (val == "1" || val == "true")
	}

	return connectorConfig{
		upstreamTimeout: upstreamTimeout,
		queues:          queues,
		rebuildInterval: rebuildInterval,
		redis_host:      redis,
		printResponse:   printResponse,
	}
}

type ResponseReceiver struct {
}

func (ResponseReceiver) Response(res types.InvokerResponse) {
	if res.Error != nil {
		log.Printf("tester got error: %s", res.Error.Error())
	} else {
		log.Printf("tester got result: [%d] %s => %s (%d) bytes", res.Status, res.Topic, res.Function, len(*res.Body))
	}
}
