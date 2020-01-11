package myfilesystem

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"kiv_zos/utils"
	"math"
	"os"
)

func (fs *MyFileSystem) SetInBitmap(value bool, bitPosition int32, bitmapAddress Address, bitmapSize Size) {
	b := fs.GetByteByBitInBitmap(bitPosition, bitmapAddress, bitmapSize)

	// which byte will be
	dstBytePosition := int(math.Floor(float64(bitPosition / 8)))
	dstBit := 7 - (bitPosition % 8)

	logrus.Infof("byte %b is going to be set at dstBit=%d", b, dstBit)

	var newByte byte
	if value {
		newByte = utils.SetBit(b, int8(dstBit))
	} else {
		newByte = utils.ClearBit(b, int8(dstBit))
	}

	fs.File.Seek(int64(bitmapAddress), os.SEEK_SET)
	fs.File.Seek(int64(dstBytePosition), os.SEEK_CUR)

	logrus.Infof("new bit %b", newByte)
	fs.File.Write([]byte{newByte})
}

func (fs *MyFileSystem) GetByteByBitInBitmap(bitPosition int32, bitmapAddress Address, bitmapSize Size) byte {
	if bitPosition >= int32(bitmapSize) {
		panic(fmt.Sprintf("Trying to set a bit in outside of a bitmap position=%d, start address=%d, bitmapSize=%d", bitPosition, bitmapAddress, bitmapSize))
	}

	_, _ = fs.File.Seek(int64(bitmapAddress), os.SEEK_SET)

	// which byte will be retrieved
	dstBytePosition := int(math.Floor(float64(bitPosition / 8)))

	fs.File.Seek(int64(dstBytePosition), os.SEEK_CUR)

	b := make([]byte, 1)

	fs.File.Read(b)

	logrus.Infof("Read byte: %b", b[0])
	return b[0]
}

func (fs *MyFileSystem) GetInBitmap(bitPosition int32, bitmapAddress Address, bitmapSize Size) bool {

	b := fs.GetByteByBitInBitmap(bitPosition, bitmapAddress, bitmapSize)

	dstBit := 7 - (bitPosition % 8)

	return utils.HasBit(b, int8(dstBit))
}
