package myfilesystem

import (
	"kiv_zos/myfilesystem"
	"testing"
)

func TestGetClusterPath(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	id, indirect := fs.GetClusterPath(0)
	if id != 0 && indirect != myfilesystem.NoIndirect {
		t.Errorf("GetClusterPath 0 failed want=%d %d, got=%d %d", 0, myfilesystem.NoIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(4)
	if id != 4 && indirect != myfilesystem.NoIndirect {
		t.Errorf("GetClusterPath 4 failed want=%d %d, got=%d %d", 4, myfilesystem.NoIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(5)
	if id != 5 && indirect != myfilesystem.FirstIndirect {
		t.Errorf("GetClusterPath 5 failed want=%d %d, got=%d %d", 6, myfilesystem.FirstIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(6)
	if id != 6 && indirect != myfilesystem.FirstIndirect {
		t.Errorf("GetClusterPath 6 failed want=%d %d, got=%d %d", 6, myfilesystem.FirstIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260)
	if id != 255 && indirect != myfilesystem.FirstIndirect {
		t.Errorf("GetClusterPath 260 failed want=%d %d, got=%d %d", 255, myfilesystem.FirstIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(261)
	if id != 0 && indirect != myfilesystem.FirstIndirect {
		t.Errorf("GetClusterPath 261 failed want=%d %d, got=%d %d", 0, myfilesystem.FirstIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(262)
	if id != 1 && indirect != myfilesystem.FirstIndirect {
		t.Errorf("GetClusterPath 262 failed want=%d %d, got=%d %d", 1, myfilesystem.FirstIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 255)
	if id != 254 && indirect != 0 {
		t.Errorf("GetClusterPath 260+255 failed want=%d %d, got=%d %d", 254, 0, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 256)
	if id != 255 && indirect != 0 {
		t.Errorf("GetClusterPath 260+256 failed want=%d %d, got=%d %d", 255, 0, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 256 + 1)
	if id != 0 && indirect != 1 {
		t.Errorf("GetClusterPath 260+256+1 failed want=%d %d, got=%d %d", 0, 1, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 256 + 2)
	if id != 1 && indirect != 1 {
		t.Errorf("GetClusterPath 260+256+2 failed want=%d %d, got=%d %d", 1, 1, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 256 + 255)
	if id != 254 && indirect != 1 {
		t.Errorf("GetClusterPath 260+256+255 failed want=%d %d, got=%d %d", 254, 1, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 256 + 256)
	if id != 255 && indirect != 1 {
		t.Errorf("GetClusterPath 260+256+256 failed want=%d %d, got=%d %d", 255, 1, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 256 + 256 + 1)
	if id != 0 && indirect != 2 {
		t.Errorf("GetClusterPath 260+256+256+1 failed want=%d %d, got=%d %d", 0, 2, id, indirect)
	}

	id, indirect = fs.GetClusterPath(5 + 256 + 256*256)
	if id != myfilesystem.FileTooLarge && indirect != myfilesystem.FileTooLarge {
		t.Errorf("GetClusterPath 5+256+256*256 failed want=%d %d, got=%d %d", myfilesystem.FileTooLarge, myfilesystem.FileTooLarge, id, indirect)
	}

	fs.Close()
}
