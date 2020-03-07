// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
	"net/http"

	"github.com/openfaas/faas/gateway/requests"
)

// FunctionLookupBuilder builds a list of OpenFaaS functions
type FunctionLookupBuilder struct {
	GatewayURL string
	Client     *http.Client
}

// Build compiles a map of topic names and functions that have
// advertised to receive messages on said topic
func (s *FunctionLookupBuilder) Build() (map[string][]string, error) {
	var err error
  log.Println("Entering Build function")
	serviceMap := make(map[string][]string)
	str := fmt.Sprintf("%s/system/functions", "http://a56d6c9b55f2011eaae4402584498c9a-350195318.us-west-2.elb.amazonaws.com:8080")
	log.Println("Req: %s",str)
	//req, _ := http.NewRequest(http.MethodGet, str, nil)
  res,reqErr := http.Get(str)
	//res, reqErr := s.Client.Do(req)

	if reqErr != nil {
    
		return serviceMap, reqErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, _ := ioutil.ReadAll(res.Body)

	functions := []requests.Function{}
	marshalErr := json.Unmarshal(bytesOut, &functions)

	if marshalErr != nil {
		return serviceMap, marshalErr
	}

	for _, function := range functions {
		if function.Annotations != nil {
			annotations := *function.Annotations

			if topic, pass := annotations["topic"]; pass {

				if serviceMap[topic] == nil {
					serviceMap[topic] = []string{}
				}
				serviceMap[topic] = append(serviceMap[topic], function.Name)
			}
		}
	}

	return serviceMap, err
}
