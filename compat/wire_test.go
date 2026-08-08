// Package compat holds tests about the wire format itself rather than about
// any one field.
//
// It is a package of its own, outside pb/, because everything in pb/ is
// generated and `make verify` fails on any difference there — a handwritten
// file living among generated ones is a file somebody eventually deletes while
// regenerating.
package compat

import (
	"testing"

	"github.com/EnvSight/envsight-api/pb"
	"google.golang.org/protobuf/proto"
)

// The two properties the presence design rests on.
//
// Every field added after v1.2.0 is `optional`, which costs a presence bit and
// buys the difference between "did not report" and "reported zero". For these
// fields zero is a reading a machine genuinely produces — no traffic, no
// steal, no swap in use — so conflating the two would draw a saturated host as
// an idle one, and would draw every not-yet-upgraded agent as a host whose
// disks and network went silent.
//
// Both directions are checked because only one of them fails loudly. A missing
// presence bit turns absent into zero, which renders as a plausible chart
// nobody questions.

// An agent built against v1.2.0 sends fields 1-8 and nothing else.
func TestOldAgentMessage_NewFieldsReadAsAbsent(t *testing.T) {
	old := &pb.Metrics{
		CpuPercent: 42.5,
		MemUsedMb:  1000,
		MemTotalMb: 8192,
		Load_1:     1.5,
	}
	wire, err := proto.Marshal(old)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got pb.Metrics
	if err := proto.Unmarshal(wire, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.CpuPercent != 42.5 || got.MemUsedMb != 1000 {
		t.Errorf("the old fields did not survive: %+v", &got)
	}

	for name, present := range map[string]bool{
		"cpu_steal_percent":    got.CpuStealPercent != nil,
		"cpu_iowait_percent":   got.CpuIowaitPercent != nil,
		"mem_available_mb":     got.MemAvailableMb != nil,
		"swap_used_mb":         got.SwapUsedMb != nil,
		"disk_fullest_percent": got.DiskFullestPercent != nil,
		"disk_fullest_mount":   got.DiskFullestMount != nil,
		"disk_inodes_percent":  got.DiskInodesPercent != nil,
		"disk_util_percent":    got.DiskUtilPercent != nil,
		"disk_await_ms":        got.DiskAwaitMs != nil,
		"disk_read_bps":        got.DiskReadBps != nil,
		"disk_write_bps":       got.DiskWriteBps != nil,
		"net_rx_bps":           got.NetRxBps != nil,
		"net_tx_bps":           got.NetTxBps != nil,
		"process_count":        got.ProcessCount != nil,
		"procs_blocked":        got.ProcsBlocked != nil,
	} {
		if present {
			t.Errorf("%s reads as present in a message that never carried it", name)
		}
	}
}

// And the other direction: a reported zero must not vanish on the way.
func TestReportedZero_SurvivesAsZero(t *testing.T) {
	zero := float64(0)
	zeroU := uint32(0)
	sent := &pb.Metrics{
		NetRxBps:        &zero,
		DiskUtilPercent: &zero,
		ProcsBlocked:    &zeroU,
	}
	wire, err := proto.Marshal(sent)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got pb.Metrics
	if err := proto.Unmarshal(wire, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for name, p := range map[string]bool{
		"net_rx_bps":        got.NetRxBps != nil,
		"disk_util_percent": got.DiskUtilPercent != nil,
		"procs_blocked":     got.ProcsBlocked != nil,
	} {
		if !p {
			t.Errorf("%s was reported as 0 and came back absent", name)
		}
	}
	if got.NetRxBps != nil && *got.NetRxBps != 0 {
		t.Errorf("net_rx_bps = %v, want 0", *got.NetRxBps)
	}
}

// A field this build has never heard of survives being read and written
// again.
//
// This is what makes the three repositories deployable independently: a server
// running v1.3.0 may receive a beat from an agent built against v1.4.0, and it
// must neither reject the message nor quietly drop the parts it cannot name.
//
// Tested with a tag no message in this contract defines, rather than by
// decoding one message as another — a stand-in that happens to work today
// because two unrelated fields collide, and would stop meaning anything the
// moment either changed.
func TestUnknownField_SurvivesAReadAndRewrite(t *testing.T) {
	known := &pb.Metrics{CpuPercent: 10}
	wire, err := proto.Marshal(known)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Field 900, varint, value 7. Well past anything defined here.
	future := append(append([]byte{}, wire...), 0xE0, 0x38, 0x07)

	var got pb.Metrics
	if err := proto.Unmarshal(future, &got); err != nil {
		t.Fatalf("a message with a newer field was rejected: %v", err)
	}
	if got.CpuPercent != 10 {
		t.Errorf("cpu_percent = %v, want 10", got.CpuPercent)
	}

	again, err := proto.Marshal(&got)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	if len(again) <= len(wire) {
		t.Error("the unknown field was dropped on rewrite rather than carried through")
	}
}

// FilesystemInfo rides on NodeState and is not a metrics field. Pinned because
// the temptation to persist it is exactly what the message comment argues
// against, and a test is harder to skim past than a comment.
func TestFilesystems_TravelOnNodeStateNotOnMetrics(t *testing.T) {
	st := &pb.NodeState{
		AgentUuid: "a",
		Filesystems: []*pb.FilesystemInfo{
			{Mountpoint: "/", TotalMb: 100, UsedMb: 40},
			{Mountpoint: "/var", TotalMb: 50, UsedMb: 48},
		},
	}
	wire, err := proto.Marshal(st)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got pb.NodeState
	if err := proto.Unmarshal(wire, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Filesystems) != 2 || got.Filesystems[1].Mountpoint != "/var" {
		t.Errorf("filesystems did not survive: %+v", got.Filesystems)
	}
}
