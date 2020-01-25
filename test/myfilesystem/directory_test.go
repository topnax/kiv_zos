package myfilesystem

import (
	"kiv_zos/myfilesystem"
	"testing"
)

func TestSimpleReadOrder(t *testing.T) {
	want := myfilesystem.ReadOrder{
		ClusterId: 0,
		Start:     0,
		Bytes:     24,
	}
	got := myfilesystem.GetReadOrder(0, 24)[0]
	if got != want {
		t.Errorf("Simple read order failed want=%v, got=%d", want, got)
	}

	want = myfilesystem.ReadOrder{
		ClusterId: 1,
		Start:     0,
		Bytes:     24,
	}

	got = myfilesystem.GetReadOrder(1024, 24)[0]
	if got != want {
		t.Errorf("Simple read order failed want=%v, got=%d", want, got)
	}

	want = myfilesystem.ReadOrder{
		ClusterId: 1,
		Start:     24,
		Bytes:     24,
	}
	got = myfilesystem.GetReadOrder(1024+24, 24)[0]
	if got != want {
		t.Errorf("Simple read order failed want=%v, got=%d", want, got)
	}
}

func TestSimpleReadOrder2(t *testing.T) {
	want := []myfilesystem.ReadOrder{{
		ClusterId: 0,
		Start:     1020,
		Bytes:     4,
	}, {
		ClusterId: 1,
		Start:     0,
		Bytes:     5,
	}}
	got := myfilesystem.GetReadOrder(1020, 9)
	if got[0] != want[0] || got[1] != want[1] {
		t.Errorf("Simple read order failed want=%v, got=%d", want, got)
	}

	want = []myfilesystem.ReadOrder{{
		ClusterId: 1,
		Start:     2000 - myfilesystem.ClusterSize,
		Bytes:     48,
	}, {
		ClusterId: 2,
		Start:     0,
		Bytes:     52,
	}}
	got = myfilesystem.GetReadOrder(2000, 100)
	if got[0] != want[0] || got[1] != want[1] {
		t.Errorf("Simple read order failed want=%v, got=%d", want, got)
	}
}
