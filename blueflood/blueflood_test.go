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
	"bytes"
	"encoding/gob"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBluefloodPlugin(t *testing.T) {
	Convey("Meta returns proper metadata", t, func() {
		meta := Meta()
		So(meta.Name, ShouldResemble, pluginName)
		So(meta.Version, ShouldResemble, pluginVersion)
		So(meta.Type, ShouldResemble, plugin.PublisherPluginType)
	})

	Convey("Create Blueflood Publisher", t, func() {
		bp := NewBluefloodPublisher()
		Convey("So Blueflood Publisher should not be nil", func() {
			So(bp, ShouldNotBeNil)
		})
		Convey("So Blueflood Publisher shoud be of type bluefloodPublisher", func() {
			So(bp, ShouldHaveSameTypeAs, &BluefloodPublisher{})
		})

		configPolicy, err := bp.GetConfigPolicy()
		Convey("GetConfigPolicy() should return a config policy", func() {
			Convey("So config policy should not be nil", func() {
				So(configPolicy, ShouldNotBeNil)
			})
			Convey("So GetConfigPolicy() should not return an error", func() {
				So(err, ShouldBeNil)
			})
			Convey("So config policy should be of cpolicy.ConfigPolicy type", func() {
				So(configPolicy, ShouldHaveSameTypeAs, &cpolicy.ConfigPolicy{})
			})
			testConfig := make(map[string]ctypes.ConfigValue)
			testConfig["server"] = ctypes.ConfigValueStr{Value: "127.0.0.1:8080"}
			testConfig["rollupNum"] = ctypes.ConfigValueInt{Value: 20}
			testConfig["ttlInSeconds"] = ctypes.ConfigValueInt{Value: 172800}
			testConfig["timeout"] = ctypes.ConfigValueInt{Value: 10}

			cfg, errs := configPolicy.Get([]string{""}).Process(testConfig)

			Convey("So config policy should process testConfig and return a config", func() {
				So(cfg, ShouldNotBeNil)
			})
			Convey("So testConfig processing should return no errors", func() {
				So(errs.HasErrors(), ShouldBeFalse)
			})
		})
	})
}

func TestBluefloodPluginMetrics(t *testing.T) {
	config := make(map[string]ctypes.ConfigValue)
	intMetrics := []plugin.MetricType{
		*plugin.NewMetricType(core.NewNamespace("/staples/test/int1"), time.Now(), nil, "int", 1),
		*plugin.NewMetricType(core.NewNamespace("/staples/test/int2"), time.Now().Add(2*time.Second), nil, "int", 2),
		*plugin.NewMetricType(core.NewNamespace("/staples/test/int3"), time.Now().Add(3*time.Second), nil, "int", 3),
	}
	floatMetrics := []plugin.MetricType{
		*plugin.NewMetricType(core.NewNamespace("/staples/test/float1"), time.Now(), nil, "float", 1.5),
		*plugin.NewMetricType(core.NewNamespace("/staples/test/float2"), time.Now().Add(2*time.Second), nil, "float", 2.5),
		*plugin.NewMetricType(core.NewNamespace("/staples/test/float3"), time.Now().Add(3*time.Second), nil, "float", 3.5),
	}
	stringIntMetrics := []plugin.MetricType{
		*plugin.NewMetricType(core.NewNamespace("/staples/test/string1"), time.Now(), nil, "string", "1"),
		*plugin.NewMetricType(core.NewNamespace("/staples/test/string2"), time.Now().Add(2*time.Second), nil, "string", "2"),
		*plugin.NewMetricType(core.NewNamespace("/staples/test/string3"), time.Now().Add(3*time.Second), nil, "string", "3"),
	}
	// stringMetrics := []plugin.MetricType{
	// 	*plugin.NewMetricType(core.NewNamespace("/staples/test/string1"), time.Now(), nil, "string", "one"),
	// 	*plugin.NewMetricType(core.NewNamespace("/staples/test/string2"), time.Now().Add(2*time.Second), nil, "string", "two"),
	// 	*plugin.NewMetricType(core.NewNamespace("/staples/test/string3"), time.Now().Add(3*time.Second), nil, "string", "three"),
	// }
	config["server"] = ctypes.ConfigValueStr{Value: "http://localhost:9090"}
	config["rollupNum"] = ctypes.ConfigValueInt{Value: 20}
	config["ttlInSeconds"] = ctypes.ConfigValueInt{Value: 172800}
	config["timeout"] = ctypes.ConfigValueInt{Value: 0}

	Convey("Test blueflood plugin and the ability to send different metrics types", t, func() {
		var buf bytes.Buffer
		buf.Reset()
		enc := gob.NewEncoder(&buf)

		bp := NewBluefloodPublisher()
		cp, _ := bp.GetConfigPolicy()
		cfg, _ := cp.Get([]string{""}).Process(config)

		Convey("Publish integer metrics", func() {
			enc.Encode(intMetrics)
			err := bp.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish float metrics", func() {
			enc.Encode(floatMetrics)
			err := bp.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish string integers metrics", func() { // Should be marshalled into json as a numeric
			enc.Encode(stringIntMetrics)
			err := bp.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		// Convey("Publish non-numeric strings metrics", func() { // Should result in no metrics being pushed
		// 	enc.Encode(stringMetrics)
		// 	err := bp.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
		// 	So(err, ShouldBeNil)
		// })

	})

}
