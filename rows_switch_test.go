// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package hamapping_test

import (
	"testing"

	onoffdef "github.com/SukramJ/go-fabric/cluster/spec/onoff"
	"github.com/SukramJ/go-fabric/controller/client"
	"github.com/SukramJ/go-fabric/fabrictest"
	hacatalog "github.com/SukramJ/go-ha-catalog"
	"github.com/SukramJ/go-hamqtt/model"
)

func TestSwitchRoundTrip(t *testing.T) {
	r := serve(t, entity("plug", hacatalog.PlatformSwitch, model.Description{Name: model.L("Plug"), DeviceClass: "outlet"}))
	r.apply("plug", model.State{Value: "ON", Available: true})
	if st, _ := onoffdef.StateFrom(r.values("plug", onoffdef.ClusterID)); !st.OnOff {
		t.Fatalf("OnOff %+v after ON", st)
	}
	if _, err := fabrictest.Invoke(r.ctx, r.pair, client.CommandAt(r.ep("plug"), onoffdef.ClusterID, onoffdef.CmdOff), nil); err != nil {
		t.Fatal(err)
	}
	expect(t, r.command(), "", "OFF")
	r.apply("plug", model.State{Available: false})
	if r.reachable("plug") {
		t.Fatal("reachable after Available false")
	}
}
