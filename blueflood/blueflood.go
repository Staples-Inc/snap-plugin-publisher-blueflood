package blueflood

/*
Copyright 2016 Staples, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

type ingestMetric struct {
	CollectionTime int64       `json:"collectionTime"`
	TTLInSeconds   int64       `json:"ttlInSeconds"`
	MetricValue    interface{} `json:"metricValue"`
	MetricName     string      `json:"metricName"`
}

// BfPublisher allows for publishing metrics to blueflood
type BfPublisher struct{}

// NewBfPublisher returns a blueflood publisher
func NewBfPublisher() BfPublisher {
	return BfPublisher{}
}

func publishMetrics(data []ingestMetric, server string, timeout int64) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling json data: %s\n", err.Error())
		return
	}

	buff := bytes.NewReader(jsonData)
	req, err := http.NewRequest("POST", server, buff)

	if err != nil {
		log.Warnf("Error creating Ingest POST request, error: %s\n", err.Error())
		return
	}

	httptimeout := time.Duration(timeout) * time.Second
	client := &http.Client{Timeout: httptimeout}

	req.Header.Set("Content-Type", "application/json")

	response, err := client.Do(req)

	if err != nil {
		log.Warnf("response failed: %v", err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Warnf("status: %v; metrics not ingested", response.StatusCode)
	} else {
		log.Debugf("status %v\n", response.StatusCode)
	}

	return
}

// Publish metrics to the configured blueflood server at the specified address
func (b BfPublisher) Publish(mts []plugin.Metric, cfg plugin.Config) error {
	server, err := cfg.GetString("server")
	if err != nil {
		log.Errorf("unable to parse blueflood server from configs")
		return err
	}
	rollUpNum, err := cfg.GetInt("rollupNum")
	if err != nil {
		log.Errorf("unable to parse blueflood rollUpNum from configs")
		return err
	}
	ttlInSeconds, err := cfg.GetInt("ttlInSeconds")
	if err != nil {
		log.Errorf("unable to parse blueflood ttlInSeconds from configs")
		return err
	}
	timeout, err := cfg.GetInt("timeout")
	if err != nil {
		log.Errorf("unable to parse blueflood timeout from configs")
		return err
	}

	data := []ingestMetric{}
	for _, m := range mts {
		//Ensure empty namespaces are not sent to blueflood
		if len(m.Namespace.Strings()) < 1 {
			continue
		}
		switch v := m.Data.(type) {
		case float64:
			if math.IsNaN(m.Data.(float64)) {
				log.Warningf("Data NaN and not serializable '%v': Type %T", m.Namespace, v)
				continue
			}
			data = append(data, ingestMetric{MetricName: Key(m.Namespace.Strings()), MetricValue: m.Data, TTLInSeconds: ttlInSeconds, CollectionTime: time.Now().Unix() * 1000})
		case float32, int, int32, int64, uint32, uint64:
			data = append(data, ingestMetric{MetricName: Key(m.Namespace.Strings()), MetricValue: m.Data, TTLInSeconds: ttlInSeconds, CollectionTime: time.Now().Unix() * 1000})
		case string:
			d, ok := strconv.ParseFloat(m.Data.(string), 64)
			if ok == nil {
				data = append(data, ingestMetric{MetricName: Key(m.Namespace.Strings()), MetricValue: d, TTLInSeconds: ttlInSeconds, CollectionTime: time.Now().Unix() * 1000})
			}
		default:
			log.Warningf("Unknown data received for metric '%v': Type %T", m.Namespace, v)
		}

		if int64(len(data)) == rollUpNum {
			go publishMetrics(data, server, timeout)
			data = []ingestMetric{}
		}
	}

	if len(data) > 0 {
		go publishMetrics(data, server, timeout)
	}

	return nil
}

func handleConfigErr(e error) {
	if e != nil {
		log.Panicf("Error: Config Policy not set correctly: %v", e)
	}
}

// Key returns a string representation of the namespace with "." joining
// the elements of the namespace.
func Key(n []string) string {
	return strings.Join(n, ".")
}

// GetConfigPolicy gathers configurations for the blueflood publisher
func (b BfPublisher) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewStringRule([]string{""}, "server", true)
	policy.AddNewIntRule([]string{""}, "rollupNum", false, plugin.SetDefaultInt(100))
	policy.AddNewIntRule([]string{""}, "ttlInSeconds", false, plugin.SetDefaultInt(172800))
	policy.AddNewIntRule([]string{""}, "timeout", false, plugin.SetDefaultInt(0))
	return *policy, nil
}
