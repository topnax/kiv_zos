package myfilesystem

import (
	"github.com/sirupsen/logrus"
	"kiv_zos/myfilesystem"
	"testing"
	"unsafe"
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

func TestNextDirItem(t *testing.T) {
	got := myfilesystem.NextDirItemIndex(myfilesystem.PseudoInode{FileSize: 0})
	want := myfilesystem.ID(0)
	if got != want {
		t.Errorf("NextDirItemIndex failed, want=%d, got=%d", want, got)
	}

	got = myfilesystem.NextDirItemIndex(myfilesystem.PseudoInode{FileSize: myfilesystem.Size(unsafe.Sizeof(myfilesystem.DirectoryItem{}))})
	want = myfilesystem.ID(1)
	if got != want {
		t.Errorf("NextDirItemIndex failed, want=%d, got=%d", want, got)
	}

	got = myfilesystem.NextDirItemIndex(myfilesystem.PseudoInode{FileSize: myfilesystem.Size(unsafe.Sizeof(myfilesystem.DirectoryItem{})) * 50})
	want = myfilesystem.ID(50)
	if got != want {
		t.Errorf("NextDirItemIndex failed, want=%d, got=%d", want, got)
	}

	got = myfilesystem.NextDirItemIndex(myfilesystem.PseudoInode{FileSize: myfilesystem.Size(unsafe.Sizeof(myfilesystem.DirectoryItem{})) * 99})
	want = myfilesystem.ID(99)
	if got != want {
		t.Errorf("NextDirItemIndex failed, want=%d, got=%d", want, got)
	}
}

func TestAddDirectoryItem(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	nodeId := fs.AddInode(myfilesystem.PseudoInode{
		IsDirectory: true,
	})

	dirItem := myfilesystem.DirectoryItem{
		NodeID: 2,
		Name:   myfilesystem.NameToDirName("Hello"),
	}
	fs.AddDirItem(dirItem, nodeId)

	dirItem2 := myfilesystem.DirectoryItem{
		NodeID: 3,
		Name:   myfilesystem.NameToDirName("Hiccup"),
	}

	fs.AddDirItem(dirItem2, nodeId)

	dirItems := fs.ReadDirItems(nodeId)

	if len(dirItems) != 2 {
		t.Errorf("Expected to read one dir item but got=%d", len(dirItems))
	}

	if dirItem != dirItems[0] {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem, dirItems[0])
	}

	if dirItem2 != dirItems[1] {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem2, dirItems[1])
	}
}

func TestAddDirectoryItems(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	nodeId := fs.AddInode(myfilesystem.PseudoInode{
		IsDirectory: true,
	})

	dirItem := myfilesystem.DirectoryItem{
		NodeID: 2,
		Name:   myfilesystem.NameToDirName("HELLO"),
	}

	dirItem2 := myfilesystem.DirectoryItem{
		NodeID: 3,
		Name:   myfilesystem.NameToDirName("Hicup"),
	}
	dirItem3 := myfilesystem.DirectoryItem{
		NodeID: 4,
		Name:   myfilesystem.NameToDirName("Rndmwrds"),
	}

	for i := 0; i < 20; i++ {
		fs.AddDirItem(dirItem, nodeId)
	}

	fs.AddDirItem(dirItem2, nodeId)
	fs.AddDirItem(dirItem3, nodeId)

	dirItems := fs.ReadDirItems(nodeId)

	if len(dirItems) != 22 {
		t.Fatalf("Expected to read 22 dir items but got=%d", len(dirItems))
	}

	if dirItem != dirItems[0] {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem, dirItems[0])
	}

	if dirItem != dirItems[1] {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem2, dirItems[1])
	}

	for i := 0; i < 20; i++ {
		if dirItems[i] != dirItem {
			t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem, dirItems[i])
		}
	}
	if dirItems[20] != dirItem2 {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem2, dirItems[20])
	}

	if dirItems[21] != dirItem3 {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem3, dirItems[21])
	}
}

func TestRemoveDirectoryItems(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")
	fs.Format(1 * 1024 * 1024)

	nodeId := fs.AddInode(myfilesystem.PseudoInode{
		IsDirectory: true,
	})

	dirItem := myfilesystem.DirectoryItem{
		NodeID: 2,
		Name:   myfilesystem.NameToDirName("Hello"),
	}

	dirItem2 := myfilesystem.DirectoryItem{
		NodeID: 3,
		Name:   myfilesystem.NameToDirName("Hiccup"),
	}
	dirItem3 := myfilesystem.DirectoryItem{
		NodeID: 4,
		Name:   myfilesystem.NameToDirName("Rndname"),
	}
	dirItem4 := myfilesystem.DirectoryItem{
		NodeID: 5,
		Name:   myfilesystem.NameToDirName("TestName"),
	}

	fs.AddDirItem(dirItem, nodeId)
	fs.AddDirItem(dirItem2, nodeId)
	fs.AddDirItem(dirItem3, nodeId)

	fs.RemoveDirItem(dirItem3, nodeId)

	items := fs.ReadDirItems(nodeId)

	if len(items) != 2 {
		t.Fatalf("The items length should be 2. got=%d", len(items))
	}

	if items[0] != dirItem {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem, items[0])
	}

	if items[1] != dirItem2 {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem2, items[1])
	}

	fs.AddDirItem(dirItem3, nodeId)

	items = fs.ReadDirItems(nodeId)

	if len(items) != 3 {
		t.Fatalf("The items length should be 3. got=%d", len(items))
	}

	if items[0] != dirItem {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem, items[0])
	}

	if items[1] != dirItem2 {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem2, items[1])
	}

	if items[2] != dirItem3 {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem3, items[2])
	}

	fs.AddDirItem(dirItem4, nodeId)
	fs.RemoveDirItem(dirItem2, nodeId)
	items = fs.ReadDirItems(nodeId)

	if len(items) != 3 {
		t.Fatalf("The items length should be 2. got=%d", len(items))
	}

	if items[0] != dirItem {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem, items[0])
	}

	if items[1] != dirItem4 {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem4, items[1])
	}

	if items[2] != dirItem3 {
		t.Errorf("Read incorrect diritem. Want=%v, got=%v", dirItem3, items[2])
	}
	fs.Close()
}

func TestListDirectoryItems(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")
	fs.Format(1 * 1024 * 1024)

	rootId := fs.AddInode(myfilesystem.PseudoInode{
		IsDirectory: true,
	})

	dirId := fs.AddInode(myfilesystem.PseudoInode{
		IsDirectory: false,
	})
	dirItem := myfilesystem.DirectoryItem{
		NodeID: dirId,
		Name:   myfilesystem.NameToDirName("text"),
	}

	dirId = fs.AddInode(myfilesystem.PseudoInode{
		IsDirectory: false,
	})
	dirItem2 := myfilesystem.DirectoryItem{
		NodeID: dirId,
		Name:   myfilesystem.NameToDirName("kniha"),
	}

	dirId = fs.AddInode(myfilesystem.PseudoInode{
		IsDirectory: true,
	})
	dirItem3 := myfilesystem.DirectoryItem{
		NodeID: dirId,
		Name:   myfilesystem.NameToDirName("pdfka"),
	}

	fs.AddDirItem(dirItem, rootId)
	fs.AddDirItem(dirItem2, rootId)
	fs.AddDirItem(dirItem3, rootId)

	//rootDirItem := myfilesystem.DirectoryItem{
	//	NodeID: rootId,
	//	Name:   myfilesystem.NameToDirName("slozka"),
	//}

	fs.ListDirectory(rootId)

	//fs.PrintInfo(fs.GetInodeAt(rootId), rootDirItem)
}
func TestFindDirPath(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")
	fs.RealMode = true
	fs.Format(1 * 1024 * 1024)

	slozkaId := fs.NewDirectory(0, "slozka", false)

	podSlozkaId := fs.NewDirectory(slozkaId, "podslozka", false)

	logrus.Infof("root %v", fs.ReadDirItems(0))
	logrus.Infof("1st %v", fs.ReadDirItems(slozkaId))
	logrus.Infof("2nd %v", fs.ReadDirItems(podSlozkaId))
	logrus.Infof("1st %d", slozkaId)
	logrus.Infof("2nd %d", podSlozkaId)

	want := "/slozka/podslozka/"
	got := fs.FindDirPath(podSlozkaId)
	if want != got {
		t.Errorf("FindDirPath failed, want=%s, got=%s", want, got)
	}

	want = "/slozka/"
	got = fs.FindDirPath(slozkaId)
	if want != got {
		t.Errorf("FindDirPath failed, want=%s, got=%s", want, got)
	}

	want = "/"
	got = fs.FindDirPath(0)
	if want != got {
		t.Errorf("FindDirPath failed, want=%s, got=%s", want, got)
	}
}
