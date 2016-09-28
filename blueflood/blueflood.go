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
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
)

type ingestMetric struct {
	CollectionTime int64       `json:"collectionTime"`
	TTLInSeconds   int         `json:"ttlInSeconds"`
	MetricValue    interface{} `json:"metricValue"`
	MetricName     string      `json:"metricName"`
}

const (
	pluginName    = "blueflood"
	pluginVersion = 1
	pluginType    = plugin.PublisherPluginType
)

// Meta information about this plugin
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(pluginName, pluginVersion, pluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}

// BluefloodPublisher allows for publishing metrics to blueflood
type BluefloodPublisher struct{}

// NewBluefloodPublisher returns a blueflood publisher
func NewBluefloodPublisher() *BluefloodPublisher {
	return &BluefloodPublisher{}
}

func publishMetrics(data []ingestMetric, server string, timeout int, logger *log.Logger) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Printf("Error marshalling json data: %s\n", err.Error())
		return
	}

	buff := bytes.NewReader(jsonData)
	req, err := http.NewRequest("POST", server, buff)

	if err != nil {
		logger.Warnf("Error creating Ingest POST request, error: %s\n", err.Error())
		return
	}

	httptimeout := time.Duration(timeout) * time.Second
	client := &http.Client{Timeout: httptimeout}

	req.Header.Set("Content-Type", "application/json")

	response, err := client.Do(req)

	if err != nil {
		logger.Warnf("response failed: %v", err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		logger.Warnf("Metrics not ingested, status: %v", response.StatusCode)
	} else {
		logger.Infof("status - %v", response.StatusCode)
	}

	return
}

// Publish metrics to the configured blueflood server at the specified address
func (b *BluefloodPublisher) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	logger := log.New()
	var metrics []plugin.MetricType
	switch contentType {
	case plugin.SnapGOBContentType:
		dec := gob.NewDecoder(bytes.NewBuffer(content))
		if err := dec.Decode(&metrics); err != nil {
			logger.Printf("Error decoding GOB: error=%v content=%v", err, content)
			return err
		}
	default:
		logger.Warnf("Error unknown content type '%v'", contentType)
		return fmt.Errorf("Unknown content type '%s'", contentType)
	}

	server := config["server"].(ctypes.ConfigValueStr).Value
	rollUpNum := config["rollupNum"].(ctypes.ConfigValueInt).Value
	ttlInSeconds := config["ttlInSeconds"].(ctypes.ConfigValueInt).Value
	timeout := config["timeout"].(ctypes.ConfigValueInt).Value

	data := []ingestMetric{}
	for _, m := range metrics {

		//Ensure empty namespaces are not sent to blueflood
		if m.Namespace().String() == "" {
			continue
		}
		switch v := m.Data().(type) {
		case float32, float64, int, int32, int64, uint32, uint64:
			data = append(data, ingestMetric{MetricName: Key(m.Namespace()), MetricValue: m.Data(), TTLInSeconds: ttlInSeconds, CollectionTime: time.Now().Unix() * 1000})
		case string:
			d, ok := strconv.ParseFloat(m.Data().(string), 64)
			if ok == nil {
				data = append(data, ingestMetric{MetricName: Key(m.Namespace()), MetricValue: d, TTLInSeconds: ttlInSeconds, CollectionTime: time.Now().Unix() * 1000})
			}
		default:
			logger.Warningf("Unknown data received for metric '%v': Type %T", m.Namespace(), v)
		}

		if len(data) == rollUpNum {
			go publishMetrics(data, server, timeout, logger)
			data = []ingestMetric{}
		}
	}

	if len(data) > 0 {
		go publishMetrics(data, server, timeout, logger)
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
func Key(n core.Namespace) string {
	return strings.Join(n.Strings(), ".")
}

// GetConfigPolicy gathers configurations for the blueflood publisher
func (b *BluefloodPublisher) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	serverName, err := cpolicy.NewStringRule("server", true)
	handleConfigErr(err)
	serverName.Description = "Blueflood host address"

	rollUpNum, err := cpolicy.NewIntegerRule("rollupNum", false, 100)
	handleConfigErr(err)
	rollUpNum.Description = "Configurable value to break up blueflood ingest requests into chunks of metrics"

	ttlInSeconds, err := cpolicy.NewIntegerRule("ttlInSeconds", false, 172800)
	handleConfigErr(err)
	ttlInSeconds.Description = "Blueflood ingest setting for number of seconds before data expires in blueflood ingest"

	timeoutVal, err := cpolicy.NewIntegerRule("timeout", false, 0)
	handleConfigErr(err)
	timeoutVal.Description = "Number of seconds to timeout out requests to the blueflood server"

	config.Add(serverName)
	config.Add(rollUpNum)
	config.Add(ttlInSeconds)
	config.Add(timeoutVal)
	cp.Add([]string{""}, config)
	return cp, nil
}
