package myfilesystem

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"kiv_zos/utils"
	"math"
)

func (fs *MyFileSystem) SetInBitmap(value bool, bitPosition int32, bitmapAddress Address, bitmapSize Size) {
	b := fs.GetByteByBitInBitmap(bitPosition, bitmapAddress, bitmapSize)

	// which byte will be
	dstBytePosition := int(math.Floor(float64(bitPosition / 8)))
	dstBit := 7 - (bitPosition % 8)

	log.Infof("byte %b is going to be set at dstBit=%d", b, dstBit)

	var newByte byte
	if value {
		newByte = utils.SetBit(b, int8(dstBit))
	} else {
		newByte = utils.ClearBit(b, int8(dstBit))
	}

	_, err := fs.File.Seek(int64(bitmapAddress), io.SeekStart)
	if err == nil {
		_, err = fs.File.Seek(int64(dstBytePosition), io.SeekCurrent)
		if err == nil {
			log.Infof("new bit %b", newByte)
			_, err = fs.File.Write([]byte{newByte})
			if err != nil {
				log.Error(err)
				panic("could not write")
			}
		} else {
			log.Error(err)
			panic("could not seek 2")
		}

	} else {
		log.Error(err)
		panic("could not seek")
	}
}

func (fs *MyFileSystem) GetByteByBitInBitmap(bitPosition int32, bitmapAddress Address, bitmapSize Size) byte {
	if bitPosition >= int32(bitmapSize) {
		panic(fmt.Sprintf("Trying to set a bit in outside of a bitmap position=%d, start address=%d, bitmapSize=%d", bitPosition, bitmapAddress, bitmapSize))
	}

	_, _ = fs.File.Seek(int64(bitmapAddress), io.SeekStart)

	// which byte will be retrieved
	dstBytePosition := int(math.Floor(float64(bitPosition / 8)))

	_, err := fs.File.Seek(int64(dstBytePosition), io.SeekCurrent)

	if err != nil {
		log.Error(err)
		panic("could not seek")
	}

	b := make([]byte, 1)

	_, err = fs.File.Read(b)
	if err != nil {
		log.Error(err)
		panic("could not read")
	}

	log.Infof("Read byte: %b", b[0])
	return b[0]
}

func (fs *MyFileSystem) GetInBitmap(bitPosition int32, bitmapAddress Address, bitmapSize Size) bool {

	b := fs.GetByteByBitInBitmap(bitPosition, bitmapAddress, bitmapSize)

	dstBit := 7 - (bitPosition % 8)

	return utils.HasBit(b, int8(dstBit))
}