package myfilesystem

import (
	"kiv_zos/myfilesystem"
	"testing"
)

func TestSuperBlockCreateAndLoad(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(5 * 1024 * 1024)

	fs.Close()

	loaded := myfilesystem.NewMyFileSystem("testfs")
	loaded.Load()

	if fs.SuperBlock.ClusterSize != loaded.SuperBlock.ClusterSize {
		t.Errorf("Loaded FS does not have the same clustersize: wanted %d, got %d", fs.SuperBlock.ClusterSize, loaded.SuperBlock.ClusterSize)
	}

	if fs.SuperBlock.ClusterCount != loaded.SuperBlock.ClusterCount {
		t.Errorf("Loaded FS does not have the same ClusterCount: wanted %d, got %d", fs.SuperBlock.ClusterCount, loaded.SuperBlock.ClusterCount)
	}

	if fs.SuperBlock.DataBitmapStartAddress != loaded.SuperBlock.DataBitmapStartAddress {
		t.Errorf("Loaded FS does not have the same DataBitmapStartAddress: wanted %d, got %d", fs.SuperBlock.DataBitmapStartAddress, loaded.SuperBlock.DataBitmapStartAddress)
	}

	if fs.SuperBlock.DataStartAddress != loaded.SuperBlock.DataStartAddress {
		t.Errorf("Loaded FS does not have the same DataStartAddress: wanted %d, got %d", fs.SuperBlock.DataStartAddress, loaded.SuperBlock.DataStartAddress)
	}

	if fs.SuperBlock.DiskSize != loaded.SuperBlock.DiskSize {
		t.Errorf("Loaded FS does not have the same DiskSize: wanted %d, got %d", fs.SuperBlock.DiskSize, loaded.SuperBlock.DiskSize)
	}

	if fs.SuperBlock.InodeBitmapStartAddress != loaded.SuperBlock.InodeBitmapStartAddress {
		t.Errorf("Loaded FS does not have the same InodeBitmapStartAddress: wanted %d, got %d", fs.SuperBlock.InodeBitmapStartAddress, loaded.SuperBlock.InodeBitmapStartAddress)
	}

	if fs.SuperBlock.InodeStartAddress != loaded.SuperBlock.InodeStartAddress {
		t.Errorf("Loaded FS does not have the same InodeStartAddress: wanted %d, got %d", fs.SuperBlock.InodeStartAddress, loaded.SuperBlock.InodeStartAddress)
	}

	if fs.SuperBlock.Signature != loaded.SuperBlock.Signature {
		t.Errorf("Loaded FS does not have the same Signature: wanted %d, got %d", fs.SuperBlock.Signature, loaded.SuperBlock.Signature)
	}

	if fs.SuperBlock.VolumeDescriptor != loaded.SuperBlock.VolumeDescriptor {
		t.Errorf("Loaded FS does not have the same VolumeDescriptor: wanted %d, got %d", fs.SuperBlock.VolumeDescriptor, loaded.SuperBlock.VolumeDescriptor)
	}

	loaded.Close()
}
