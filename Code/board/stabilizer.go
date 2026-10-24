package board

type BoardStabilizer struct {
	stableBoard   [8][8]string
	candidates    [8][8]string
	counts        [8][8]int
	requiredCount int
}

func NewBoardStabilizer(req int) *BoardStabilizer {
	if req < 1 {
		req = 1
	}
	s := &BoardStabilizer{requiredCount: req}
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			s.stableBoard[r][c] = "."
		}
	}
	return s
}

func (s *BoardStabilizer) Reset() {
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			s.stableBoard[r][c] = "."
			s.candidates[r][c] = ""
			s.counts[r][c] = 0
		}
	}
}

func (s *BoardStabilizer) Filter(raw [8][8]string) [8][8]string {
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			det := raw[r][c]

			if det == s.stableBoard[r][c] {
				s.counts[r][c] = 0
				continue
			}

			if det == s.candidates[r][c] {
				s.counts[r][c]++
				if s.counts[r][c] >= s.requiredCount {
					s.stableBoard[r][c] = det
					s.counts[r][c] = 0
				}
			} else {
				s.candidates[r][c] = det
				s.counts[r][c] = 1
				if s.requiredCount == 1 {
					s.stableBoard[r][c] = det
				}
			}
		}
	}
	return s.stableBoard
}
