// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package hamapping_test

import (
	"testing"

	rvcopdef "github.com/SukramJ/go-fabric/cluster/spec/rvcoperationalstate"
	rvcrunmodedef "github.com/SukramJ/go-fabric/cluster/spec/rvcrunmode"
	"github.com/SukramJ/go-fabric/controller/client"
	"github.com/SukramJ/go-fabric/device"
	"github.com/SukramJ/go-fabric/eligibility"
	"github.com/SukramJ/go-fabric/fabrictest"
	hacatalog "github.com/SukramJ/go-ha-catalog"
	"github.com/SukramJ/go-hamqtt/model"

	"github.com/SukramJ/go-hamapping"
)

func TestVacuum(t *testing.T) {
	e := entity("vac", hacatalog.PlatformVacuum, model.Description{Extra: map[string]any{
		"supported_features": []string{"START", "PAUSE", "RETURN_HOME", "STATE"},
	}})
	r := serve(t, e)
	wantType(t, r, "vac", &device.RoboticVacuumCleanerType, eligibility.StatePartial)
	r.apply("vac", value("cleaning"))
	run, _ := rvcrunmodedef.StateFrom(r.values("vac", rvcrunmodedef.ClusterID))
	if run.CurrentMode != 1 || len(run.SupportedModes) != 2 || run.SupportedModes[0].ModeTags[0].Value != rvcrunmodedef.ModeTagIdle {
		t.Fatalf("run mode %+v", run)
	}
	op, _ := rvcopdef.StateFrom(r.values("vac", rvcopdef.ClusterID))
	if op.OperationalState != rvcopdef.OperationalStateRunning || len(op.OperationalStateList) != 7 ||
		op.OperationalError.ErrorStateId != rvcopdef.ErrorStateNoError {
		t.Fatalf("operational state %+v", op)
	}
	resp, err := fabrictest.Invoke(r.ctx, r.pair, client.CommandAt(r.ep("vac"), rvcrunmodedef.ClusterID, rvcrunmodedef.CmdChangeToMode),
		rvcrunmodedef.ChangeToModeRequest{NewMode: 0})
	if err != nil {
		t.Fatal(err)
	}
	expect(t, r.command(), hamapping.MethodRunMode, "0")
	if got := changeToModeStatus(t, resp); got != rvcrunmodedef.ModeChangeStatusSuccess {
		t.Fatalf("ChangeToMode 0: status %v", got)
	}
	resp, err = fabrictest.Invoke(r.ctx, r.pair, client.CommandAt(r.ep("vac"), rvcrunmodedef.ClusterID, rvcrunmodedef.CmdChangeToMode),
		rvcrunmodedef.ChangeToModeRequest{NewMode: 7})
	if err != nil {
		t.Fatal(err)
	}
	if got := changeToModeStatus(t, resp); got != rvcrunmodedef.ModeChangeStatusUnsupportedMode {
		t.Fatalf("ChangeToMode 7: status %v, want UnsupportedMode", got)
	}
	r.invoke("vac", rvcopdef.ClusterID, rvcopdef.CmdPause, nil)
	expect(t, r.command(), hamapping.MethodPause, "PAUSE")
	r.invoke("vac", rvcopdef.ClusterID, rvcopdef.CmdResume, nil)
	expect(t, r.command(), hamapping.MethodResume, "START")
	r.invoke("vac", rvcopdef.ClusterID, rvcopdef.CmdGoHome, nil)
	expect(t, r.command(), hamapping.MethodGoHome, "RETURN_TO_BASE")
	r.apply("vac", value("docked"))
	if op, _ := rvcopdef.StateFrom(r.values("vac", rvcopdef.ClusterID)); op.OperationalState != rvcopdef.OperationalStateDocked {
		t.Fatalf("docked: %v", op.OperationalState)
	}
	r.noMore()
}

func TestVacuumWithoutStart(t *testing.T) {
	_, v := hamapping.Map(testDevice, []model.Entity{entity("vac", hacatalog.PlatformVacuum, model.Description{})})
	if v[0].Row != "vacuum/unmapped" || v[0].State != eligibility.StateUnmappable {
		t.Fatalf("verdict %+v", v[0])
	}
}

func changeToModeStatus(t *testing.T, res fabrictest.Result) rvcrunmodedef.ModeChangeStatus {
	t.Helper()
	var resp rvcrunmodedef.ChangeToModeResponse
	if err := resp.DecodeTLV(res.Fields); err != nil {
		t.Fatal(err)
	}
	return resp.Status
}
