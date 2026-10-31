// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gen

import (
	"reflect"

	booleanstatedef "github.com/SukramJ/go-fabric/cluster/spec/booleanstate"
	colorcontroldef "github.com/SukramJ/go-fabric/cluster/spec/colorcontrol"
	doorlockdef "github.com/SukramJ/go-fabric/cluster/spec/doorlock"
	electricalenergymeasurementdef "github.com/SukramJ/go-fabric/cluster/spec/electricalenergymeasurement"
	electricalpowermeasurementdef "github.com/SukramJ/go-fabric/cluster/spec/electricalpowermeasurement"
	fancontroldef "github.com/SukramJ/go-fabric/cluster/spec/fancontrol"
	illuminancemeasurementdef "github.com/SukramJ/go-fabric/cluster/spec/illuminancemeasurement"
	levelcontroldef "github.com/SukramJ/go-fabric/cluster/spec/levelcontrol"
	occupancysensingdef "github.com/SukramJ/go-fabric/cluster/spec/occupancysensing"
	onoffdef "github.com/SukramJ/go-fabric/cluster/spec/onoff"
	powersourcedef "github.com/SukramJ/go-fabric/cluster/spec/powersource"
	powertopologydef "github.com/SukramJ/go-fabric/cluster/spec/powertopology"
	pressuremeasurementdef "github.com/SukramJ/go-fabric/cluster/spec/pressuremeasurement"
	relativehumiditymeasurementdef "github.com/SukramJ/go-fabric/cluster/spec/relativehumiditymeasurement"
	smokecoalarmdef "github.com/SukramJ/go-fabric/cluster/spec/smokecoalarm"
	temperaturemeasurementdef "github.com/SukramJ/go-fabric/cluster/spec/temperaturemeasurement"
	thermostatdef "github.com/SukramJ/go-fabric/cluster/spec/thermostat"
	windowcoveringdef "github.com/SukramJ/go-fabric/cluster/spec/windowcovering"
)

// def is what the generator knows of one go-fabric <cluster>def package:
// its import path and alias, the cluster's name and id, and the types it
// reads field names and field types from.
type def struct {
	name    string
	id      uint32
	path    string
	alias   string
	declare reflect.Type
	patch   reflect.Type
	state   reflect.Type
	// server is the typed server (*<cluster>def.Server), whose
	// Emit<Event> methods name the cluster's events and their data.
	server reflect.Type
}

const specPath = "github.com/SukramJ/go-fabric/cluster/spec/"

func newDef[D, P, S, V any](name string, id uint32, pkg string) def {
	return def{
		name: name, id: id, path: specPath + pkg, alias: pkg + "def",
		declare: reflect.TypeFor[D](), patch: reflect.TypeFor[P](), state: reflect.TypeFor[S](),
		server: reflect.TypeFor[*V](),
	}
}

// defs are the cluster packages rows may name. A row naming another
// cluster fails generation until its package is listed here.
var defs = func() map[string]def {
	all := []def{
		newDef[booleanstatedef.Declare, booleanstatedef.Patch, booleanstatedef.State, booleanstatedef.Server](booleanstatedef.Name, booleanstatedef.ClusterID, "booleanstate"),
		newDef[colorcontroldef.Declare, colorcontroldef.Patch, colorcontroldef.State, colorcontroldef.Server](colorcontroldef.Name, colorcontroldef.ClusterID, "colorcontrol"),
		newDef[doorlockdef.Declare, doorlockdef.Patch, doorlockdef.State, doorlockdef.Server](doorlockdef.Name, doorlockdef.ClusterID, "doorlock"),
		newDef[electricalenergymeasurementdef.Declare, electricalenergymeasurementdef.Patch, electricalenergymeasurementdef.State, electricalenergymeasurementdef.Server](electricalenergymeasurementdef.Name, electricalenergymeasurementdef.ClusterID, "electricalenergymeasurement"),
		newDef[electricalpowermeasurementdef.Declare, electricalpowermeasurementdef.Patch, electricalpowermeasurementdef.State, electricalpowermeasurementdef.Server](electricalpowermeasurementdef.Name, electricalpowermeasurementdef.ClusterID, "electricalpowermeasurement"),
		newDef[fancontroldef.Declare, fancontroldef.Patch, fancontroldef.State, fancontroldef.Server](fancontroldef.Name, fancontroldef.ClusterID, "fancontrol"),
		newDef[illuminancemeasurementdef.Declare, illuminancemeasurementdef.Patch, illuminancemeasurementdef.State, illuminancemeasurementdef.Server](illuminancemeasurementdef.Name, illuminancemeasurementdef.ClusterID, "illuminancemeasurement"),
		newDef[levelcontroldef.Declare, levelcontroldef.Patch, levelcontroldef.State, levelcontroldef.Server](levelcontroldef.Name, levelcontroldef.ClusterID, "levelcontrol"),
		newDef[occupancysensingdef.Declare, occupancysensingdef.Patch, occupancysensingdef.State, occupancysensingdef.Server](occupancysensingdef.Name, occupancysensingdef.ClusterID, "occupancysensing"),
		newDef[onoffdef.Declare, onoffdef.Patch, onoffdef.State, onoffdef.Server](onoffdef.Name, onoffdef.ClusterID, "onoff"),
		newDef[powersourcedef.Declare, powersourcedef.Patch, powersourcedef.State, powersourcedef.Server](powersourcedef.Name, powersourcedef.ClusterID, "powersource"),
		newDef[powertopologydef.Declare, powertopologydef.Patch, powertopologydef.State, powertopologydef.Server](powertopologydef.Name, powertopologydef.ClusterID, "powertopology"),
		newDef[pressuremeasurementdef.Declare, pressuremeasurementdef.Patch, pressuremeasurementdef.State, pressuremeasurementdef.Server](pressuremeasurementdef.Name, pressuremeasurementdef.ClusterID, "pressuremeasurement"),
		newDef[relativehumiditymeasurementdef.Declare, relativehumiditymeasurementdef.Patch, relativehumiditymeasurementdef.State, relativehumiditymeasurementdef.Server](relativehumiditymeasurementdef.Name, relativehumiditymeasurementdef.ClusterID, "relativehumiditymeasurement"),
		newDef[smokecoalarmdef.Declare, smokecoalarmdef.Patch, smokecoalarmdef.State, smokecoalarmdef.Server](smokecoalarmdef.Name, smokecoalarmdef.ClusterID, "smokecoalarm"),
		newDef[temperaturemeasurementdef.Declare, temperaturemeasurementdef.Patch, temperaturemeasurementdef.State, temperaturemeasurementdef.Server](temperaturemeasurementdef.Name, temperaturemeasurementdef.ClusterID, "temperaturemeasurement"),
		newDef[thermostatdef.Declare, thermostatdef.Patch, thermostatdef.State, thermostatdef.Server](thermostatdef.Name, thermostatdef.ClusterID, "thermostat"),
		newDef[windowcoveringdef.Declare, windowcoveringdef.Patch, windowcoveringdef.State, windowcoveringdef.Server](windowcoveringdef.Name, windowcoveringdef.ClusterID, "windowcovering"),
	}
	m := make(map[string]def, len(all))
	for _, d := range all {
		m[d.name] = d
	}
	return m
}()

// aliasOf is the alias of a def package by import path.
func aliasOf(path string) (string, bool) {
	for _, d := range defs {
		if d.path == path {
			return d.alias, true
		}
	}
	return "", false
}
