package myfilesystem

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"kiv_zos/utils"
	"math"
)

const (
	ReadSize = 4000
)

func (fs *MyFileSystem) SetInBitmap(value bool, bitPosition int32, bitmapAddress Address, bitmapSize Size) {
	b := fs.GetByteByBitInBitmap(bitPosition, bitmapAddress, bitmapSize)

	// which byte will be
	dstBytePosition := int(math.Floor(float64(bitPosition / 8)))
	dstBit := 7 - (bitPosition % 8)

	log.Debugf("byte %b is going to be set at dstBit=%d", b, dstBit)

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
			log.Debugf("new bit %b", newByte)
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
	if bitPosition >= int32(bitmapSize*8) {
		panic(fmt.Sprintf("Trying to set a bit in outside of a bitmap position=%d, Start address=%d, bitmapSize=%d", bitPosition, bitmapAddress, bitmapSize))
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

	log.Debugf("Read byte: %b", b[0])
	return b[0]
}

func (fs *MyFileSystem) GetInBitmap(bitPosition int32, bitmapAddress Address, bitmapSize Size) bool {
	b := fs.GetByteByBitInBitmap(bitPosition, bitmapAddress, bitmapSize)

	dstBit := 7 - (bitPosition % 8)

	return utils.HasBit(b, int8(dstBit))
}

func (fs *MyFileSystem) FindFreeBitInBitmap(bitmapAddress Address, length Size) ID {
	id := ID(0)
	bytes := make([]byte, 1)
	_, err := fs.File.Seek(int64(bitmapAddress), io.SeekStart)

	if err != nil {
		log.Error(err)
		panic("Could not seek")
	}
	for Size(id) < length {
		_, _ = fs.File.Read(bytes)
		for index := int8(0); index < 8; index++ {
			if !utils.HasBit(bytes[0], 7-index) {
				return id
			}
			id++
			if Size(id) >= length {
				return -1
			}
		}
	}
	log.Warnf("Free bit in bitmap not found not found")
	return -1
}

func (fs *MyFileSystem) FindFreeBitsInBitmap(desired int, bitmapAddress Address, bitmapSize Size, bitCount Size) []ID {
	if desired == -1 {
		desired = ReadSize * 8
	}
	_, err := fs.File.Seek(int64(bitmapAddress), io.SeekStart)

	ids := []ID{}

	if err != nil {
		log.Error(err)
		panic("Could not seek")
	}

	currentAddr := bitmapAddress
	i := 0
	for currentAddr < bitmapAddress+Address(bitmapSize) {
		bytes := make([]byte, utils.Min(ReadSize, int(bitmapSize)))

		read, err := fs.File.Read(bytes)
		if err != nil {
			log.Error(err)
			log.Errorf("Could not read %d Bytes", len(bytes))
			panic("Could not read")
		}

		ids = append(ids, FindFreeBitsInBytes(ID(utils.Min(desired-len(ids), read*8)), bytes, i*ReadSize*8, int(bitCount))...)

		idsLen := len(ids)
		if len(ids) >= desired {
			for i := 0; i < utils.Min(8, idsLen); i++ {
				// to be removed
				tbr := idsLen - i - 1
				if ids[tbr] >= ID(bitCount) {
					ids = append(ids[:tbr], ids[tbr+1:]...)
					log.Infof("Removing free bit that exceeded bit at tbr=%d", tbr)
				}
			}
			return ids
		}
		currentAddr += Address(len(bytes))
		i++
	}

	return ids
}

func FindFreeBitsInBytes(desired ID, bytes []byte, offset int, bitCount int) []ID {
	ids := []ID{}
	found := ID(0)
	id := ID(0)

	for _, b := range bytes {
		for index := int8(0); index < 8; index++ {
			if !utils.HasBit(b, 7-index) {
				found++
				if id+ID(offset) < ID(bitCount) {
					ids = append(ids, id+ID(offset))
				}
			}
			id++
			if found >= desired {
				return ids
			}
		}
	}

	return ids
}
