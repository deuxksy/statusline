package tray

import (
	"encoding/binary"
	"strconv"
	"strings"
)

const emptyIconLabel = "–"

// IconLabel은 요약에서 아이콘에 들어갈 대표 잔여율을 고른다.
func IconLabel(title string) string {
	quota, _, _ := strings.Cut(title, " ")
	if len(quota) < 2 || (quota[0] != 'z' && quota[0] != 'c') {
		return emptyIconLabel
	}
	number := quota[1:]
	for _, digit := range number {
		if digit < '0' || digit > '9' {
			return emptyIconLabel
		}
	}
	percent, err := strconv.Atoi(number)
	if err != nil || percent > 100 {
		return emptyIconLabel
	}
	return strconv.Itoa(percent)
}

// 각 행의 하위 3비트가 왼쪽에서 오른쪽으로 이어지는 픽셀이다.
var iconDigits = [10][5]uint8{
	{7, 5, 5, 5, 7}, {2, 6, 2, 2, 7}, {7, 1, 7, 4, 7}, {7, 1, 7, 1, 7}, {5, 5, 7, 1, 1},
	{7, 4, 7, 1, 7}, {7, 4, 7, 5, 7}, {7, 1, 1, 1, 1}, {7, 5, 7, 5, 7}, {7, 5, 7, 1, 7},
}

// Icon은 IconLabel이 반환하는 숫자 또는 대시를 16px·32px ICO로 인코딩한다.
func Icon(label string) []byte {
	const headerSize, entrySize, dibHeaderSize = 6, 16, 40
	order := binary.LittleEndian
	ico := make([]byte, headerSize+2*entrySize)
	order.PutUint16(ico[2:], 1) // ICO 유형
	order.PutUint16(ico[4:], 2)
	for i, size := range []int{16, 32} {
		pixelSize := size * size * 4
		// ICO는 BGRA 이미지 뒤에 32비트 단위로 정렬된 AND 마스크를 붙인다.
		maskSize := ((size + 31) / 32) * 4 * size
		offset := len(ico)
		length := dibHeaderSize + pixelSize + maskSize
		entry := ico[headerSize+i*entrySize:]
		entry[0], entry[1] = byte(size), byte(size)
		order.PutUint16(entry[4:], 1)
		order.PutUint16(entry[6:], 32)
		order.PutUint32(entry[8:], uint32(length))
		order.PutUint32(entry[12:], uint32(offset))

		ico = append(ico, make([]byte, length)...)
		dib := ico[offset:]
		order.PutUint32(dib, dibHeaderSize)
		order.PutUint32(dib[4:], uint32(size))
		order.PutUint32(dib[8:], uint32(size*2)) // 이미지와 마스크의 합산 높이
		order.PutUint16(dib[12:], 1)
		order.PutUint16(dib[14:], 32)
		paintIcon(dib[dibHeaderSize:dibHeaderSize+pixelSize], size, label)
	}
	return ico
}

func paintIcon(pixels []byte, size int, label string) {
	const background = 0xff633321 // BGRA: 짙은 파란색 배경으로 흰색 숫자와 대비한다.
	for i := 0; i < len(pixels); i += 4 {
		binary.LittleEndian.PutUint32(pixels[i:], background)
	}
	scale := size / 16
	scaleX, scaleY := 2*scale, 2*scale
	glyphCount := len(label)
	if label == emptyIconLabel {
		glyphCount = 1
	} else if glyphCount == 3 {
		scaleX = scale
	}
	step := 3*scaleX + scale
	left := (size - (glyphCount*step - scale)) / 2
	top := (size - 5*scaleY) / 2
	for i, digit := range label {
		glyph := [5]uint8{0, 0, 7, 0, 0}
		if digit >= '0' && digit <= '9' {
			glyph = iconDigits[digit-'0']
		}
		for y, bits := range glyph {
			for x := 0; x < 3; x++ {
				if bits&(4>>x) == 0 {
					continue
				}
				for row := range scaleY {
					for col := range scaleX {
						px, py := left+i*step+x*scaleX+col, top+y*scaleY+row
						offset := ((size-1-py)*size + px) * 4 // DIB 행은 아래부터 저장한다.
						binary.LittleEndian.PutUint32(pixels[offset:], 0xffffffff)
					}
				}
			}
		}
	}
}
