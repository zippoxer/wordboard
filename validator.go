package main

type Validator struct {
	b *Board
}

func NewValidator(b *Board) *Validator {
	return &Validator{
		b: b,
	}
}

func (v *Validator) Validate(wordSet WordSet) bool {
	for _, word := range wordSet {
		// Check all instances of the first letter of our word against
		// the board's characters.
		for x := Unit(0); x < v.b.Width; x++ {
			for y := Unit(0); y < v.b.Height; y++ {
				if v.b.Data[x][y] != word.Chars[0] {
					continue
				}

				// Check the whole word for collision at this position in the board.
				p := Point{x, y}
				if v.checkCollision(word.Chars[1:], word.Path, p, false) {
					return false
				}
			}
		}
	}

	return true
}

func (v *Validator) checkCollision(word Chars, path Path, at Point, outOfPath bool) bool {
	if len(word) == 0 {
		return outOfPath
	}

	// Define a range of (-1, -1) to (+1, +1) around the given point.
	x1 := at.X - 1
	if x1 < 0 {
		x1 = 0
	}
	y1 := at.Y - 1
	if y1 < 0 {
		y1 = 0
	}
	x2 := at.X + 1
	if x2 > v.b.Width-1 {
		x2 = v.b.Width - 1
	}
	y2 := at.Y + 1
	if y2 > v.b.Height-1 {
		y2 = v.b.Height - 1
	}

	// Check if the next letter is within range.
	for x := x1; x <= x2; x++ {
		for y := y1; y <= y2; y++ {
			p := Point{x, y}
			if v.b.At(p) == word[0] {
				if path.Index(p) == -1 {
					outOfPath = true
				}

				// Check next letter.
				if v.checkCollision(word[1:], path, Point{x, y}, outOfPath) {
					return true
				}
			}
		}
	}

	return false
}
