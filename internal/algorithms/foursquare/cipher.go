package foursquare

import (
	"errors"
	"strings"
	"unicode"
)

type FoursquareCipher struct {
	grid1 [5][5]rune
	grid2 [5][5]rune
	grid3 [5][5]rune
	grid4 [5][5]rune
}

func NewCipher(key1, key2 string) (*FoursquareCipher, error) {
	f := &FoursquareCipher{}

	if err := f.generateGrids(key1, key2); err != nil {
		return nil, err
	}

	return f, nil
}

func (f *FoursquareCipher) Encrypt(plaintext string) (string, error) {

	processed := f.prepareText(plaintext)

	if len(processed)%2 != 0 {
		processed += "X"
	}

	var result strings.Builder

	for i := 0; i < len(processed); i += 2 {
		a := rune(processed[i])
		b := rune(processed[i+1])

		row1, col1 := f.findPosition(a, f.grid1)
		row2, col2 := f.findPosition(b, f.grid4)

		encA := f.grid2[row1][col2]
		encB := f.grid3[row2][col1]

		result.WriteRune(encA)
		result.WriteRune(encB)
	}

	return result.String(), nil
}

func (f *FoursquareCipher) Decrypt(ciphertext string) (string, error) {
	if len(ciphertext)%2 != 0 {
		return "", errors.New("ciphertext must have even length")
	}

	var result strings.Builder

	for i := 0; i < len(ciphertext); i += 2 {
		a := rune(ciphertext[i])
		b := rune(ciphertext[i+1])

		row1, col1 := f.findPosition(a, f.grid2)
		row2, col2 := f.findPosition(b, f.grid3)

		decA := f.grid1[row1][col2]
		decB := f.grid4[row2][col1]

		result.WriteRune(decA)
		result.WriteRune(decB)
	}

	return result.String(), nil
}

func (f *FoursquareCipher) generateGrids(key1, key2 string) error {

	alphabet := "ABCDEFGHIKLMNOPQRSTUVWXYZ"
	f.grid1 = f.createGrid(alphabet)

	f.grid4 = f.createGrid(alphabet)

	f.grid2 = f.createGrid(f.prepareKey(key1) + alphabet)

	f.grid3 = f.createGrid(f.prepareKey(key2) + alphabet)

	return nil
}

func (f *FoursquareCipher) createGrid(chars string) [5][5]rune {
	var grid [5][5]rune
	used := make(map[rune]bool)
	idx := 0

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			for idx < len(chars) {
				ch := rune(chars[idx])
				idx++

				if !used[ch] {
					grid[i][j] = ch
					used[ch] = true
					break
				}
			}
		}
	}

	return grid
}

func (f *FoursquareCipher) findPosition(ch rune, grid [5][5]rune) (int, int) {
	ch = unicode.ToUpper(ch)

	if ch == 'J' {
		ch = 'I'
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if grid[i][j] == ch {
				return i, j
			}
		}
	}

	return -1, -1
}

func (f *FoursquareCipher) prepareKey(key string) string {
	var result strings.Builder
	used := make(map[rune]bool)

	for _, ch := range key {
		ch = unicode.ToUpper(ch)
		if ch < 'A' || ch > 'Z' {
			continue
		}

		if ch == 'J' {
			ch = 'I'
		}

		if !used[ch] {
			result.WriteRune(ch)
			used[ch] = true
		}
	}

	return result.String()
}

func (f *FoursquareCipher) prepareText(text string) string {
	var result strings.Builder

	for _, ch := range text {
		ch = unicode.ToUpper(ch)
		if ch >= 'A' && ch <= 'Z' {

			if ch == 'J' {
				ch = 'I'
			}
			result.WriteRune(ch)
		}
	}

	return result.String()
}

func (f *FoursquareCipher) GetGrids() ([5][5]rune, [5][5]rune, [5][5]rune, [5][5]rune) {
	return f.grid1, f.grid2, f.grid3, f.grid4
}
