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
package blueflood

import (
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBluefloodPlugin(t *testing.T) {
	Convey("Create Blueflood Publisher", t, func() {
		bp := NewBfPublisher()
		Convey("So Blueflood Publisher shoud be of type bluefloodPublisher", func() {
			So(bp, ShouldHaveSameTypeAs, BfPublisher{})
		})

		configPolicy, err := bp.GetConfigPolicy()
		Convey("GetConfigPolicy() should return a config policy", func() {
			Convey("So GetConfigPolicy() should not return an error", func() {
				So(err, ShouldBeNil)
			})
			Convey("So config policy should be of plugin.ConfigPolicy type", func() {
				So(configPolicy, ShouldHaveSameTypeAs, plugin.ConfigPolicy{})
			})
		})
	})
}

func TestBluefloodPluginMetrics(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost:9090",
		httpmock.NewStringResponder(200, ""))

	intMetrics := []plugin.Metric{
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "int1"), Timestamp: time.Now(), Data: 1},
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "int1"), Timestamp: time.Now(), Data: 1},
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "int2"), Timestamp: time.Now().Add(2 * time.Second), Data: 2},
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "int3"), Timestamp: time.Now().Add(3 * time.Second), Data: 3},
	}
	floatMetrics := []plugin.Metric{
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "float1"), Timestamp: time.Now(), Data: 1.5},
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "float2"), Timestamp: time.Now().Add(2 * time.Second), Data: 2.5},
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "float3"), Timestamp: time.Now().Add(3 * time.Second), Data: 3.5},
	}
	uint64Metrics := []plugin.Metric{
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "float1"), Timestamp: time.Now(), Data: uint64(15)},
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "float2"), Timestamp: time.Now().Add(2 * time.Second), Data: uint64(25)},
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "float3"), Timestamp: time.Now().Add(3 * time.Second), Data: uint64(35)},
	}
	stringIntMetrics := []plugin.Metric{
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "string1"), Timestamp: time.Now(), Data: "1"},
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "string2"), Timestamp: time.Now().Add(2 * time.Second), Data: "2"},
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "string3"), Timestamp: time.Now().Add(3 * time.Second), Data: "3"},
	}
	stringMetrics := []plugin.Metric{
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "string1"), Timestamp: time.Now(), Data: "one"},
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "string2"), Timestamp: time.Now().Add(2 * time.Second), Data: "two"},
		plugin.Metric{Namespace: plugin.NewNamespace("staples", "test", "string3"), Timestamp: time.Now().Add(3 * time.Second), Data: "three"},
	}
	testConfig := plugin.Config{
		"server":       "http://localhost:9090",
		"rollupNum":    int64(20),
		"ttlInSeconds": int64(172800),
		"timeout":      int64(10),
	}

	Convey("Test blueflood plugin and the ability to send different metrics types", t, func() {
		bp := NewBfPublisher()
		Convey("Publish integer metrics", func() {
			err := bp.Publish(intMetrics, testConfig)
			So(err, ShouldBeNil)
		})
		Convey("Publish float metrics", func() {
			err := bp.Publish(floatMetrics, testConfig)
			So(err, ShouldBeNil)
		})
		Convey("Publish string integers metrics", func() { // Should be marshalled into json as a numeric
			err := bp.Publish(stringIntMetrics, testConfig)
			So(err, ShouldBeNil)
		})
		Convey("Publish non-numeric strings metrics", func() { // Should result in no metrics being pushed
			err := bp.Publish(stringMetrics, testConfig)
			So(err, ShouldBeNil)
		})
		Convey("Publish uint64 metrics", func() {
			err := bp.Publish(uint64Metrics, testConfig)
			So(err, ShouldBeNil)
		})
	})

}
