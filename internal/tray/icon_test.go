package tray

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestIconLabel(t *testing.T) {
	for title, want := range map[string]string{
		"z87 w45": "z87", "c48": "c48", "z0": "z0", "c100": "c100", "z100 w100": "z100",
		"…": "z–", "c-": "c–", "z101": "–", "c+4": "–", "bad": "–", "": "–",
	} {
		if got := IconLabel(title); got != want {
			t.Errorf("IconLabel(%q) = %q, want %q", title, got, want)
		}
	}
}

func TestIcon(t *testing.T) {
	if bytes.Equal(Icon("z48"), Icon("c48")) {
		t.Error("provider glyphs must render differently at the same quota")
	}
	for _, label := range []string{"z0", "z100", "c100", "c–", "z–", "–"} {
		t.Run(label, func(t *testing.T) {
			ico := Icon(label)
			if len(ico) < 38 || !bytes.Equal(ico[:6], []byte{0, 0, 1, 0, 2, 0}) {
				t.Fatal("invalid ICO directory")
			}
			nextOffset := 38
			for i, size := range []int{16, 32} {
				entry := ico[6+i*16:][:16]
				length := int(binary.LittleEndian.Uint32(entry[8:]))
				offset := int(binary.LittleEndian.Uint32(entry[12:]))
				if int(entry[0]) != size || int(entry[1]) != size || offset != nextOffset ||
					length != 40+size*size*4+size*4 || offset+length > len(ico) {
					t.Fatalf("invalid %dx%d ICO entry: %v", size, size, entry)
				}
				nextOffset += length
				dib := ico[offset:nextOffset]
				if binary.LittleEndian.Uint32(dib) != 40 || binary.LittleEndian.Uint32(dib[4:]) != uint32(size) ||
					binary.LittleEndian.Uint32(dib[8:]) != uint32(2*size) || binary.LittleEndian.Uint16(dib[14:]) != 32 {
					t.Fatalf("invalid %dx%d DIB header", size, size)
				}
				columns, rows := make([]bool, size), make([]bool, size)
				for y := 0; y < size; y++ {
					for x := 0; x < size; x++ {
						if pixel := dib[40+4*(y*size+x):][:4]; bytes.Equal(pixel, []byte{255, 255, 255, 255}) {
							columns[x], rows[y] = true, true
						}
					}
				}
				groups, height := 0, 0
				for x, marked := range columns {
					if marked && (x == 0 || !columns[x-1]) {
						groups++
					}
				}
				for _, marked := range rows {
					if marked {
						height++
					}
				}
				wantGroups, wantHeight := len([]rune(label)), 10*size/16
				if label == "–" {
					wantHeight = 2 * size / 16
				} else if label == "z–" || label == "c–" {
					wantHeight = 6 * size / 16
				}
				if groups != wantGroups || height != wantHeight {
					t.Errorf("%dpx icon: %d glyphs, %dpx tall; want %d glyphs, %dpx tall", size, groups, height, wantGroups, wantHeight)
				}
			}
			if nextOffset != len(ico) {
				t.Fatal("unexpected data after the last image")
			}
		})
	}
}
