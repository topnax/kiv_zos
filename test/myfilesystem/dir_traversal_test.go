package myfilesystem

import (
	"kiv_zos/myfilesystem"
	"testing"
)

func TestGetDirNames(t *testing.T) {
	want := []string{"root", "slozka", "podslozka"}
	got := myfilesystem.GetDirNames("root/slozka/podslozka")

	if len(want) != len(got) {
		t.Fatalf("Incorrect length, want=%d, got=%d", len(want), len(got))
	}

	for index, item := range got {
		if item != want[index] {
			t.Errorf("Incorrect item, want=%s, got=%s", want[index], item)
		}
	}

	want = []string{"/", "root", "slozka", "podslozka"}
	got = myfilesystem.GetDirNames("/root/slozka/podslozka")

	if len(want) != len(got) {
		t.Fatalf("Incorrect length, want=%d, got=%d", len(want), len(got))
	}

	for index, item := range got {
		if item != want[index] {
			t.Errorf("Incorrect item, want=%s, got=%s", want[index], item)
		}
	}
}
