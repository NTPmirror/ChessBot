package chesscom

import (
	"image"
	"math"
)

// PopupAction represents an action required to dismiss, confirm, or scroll over a dialog/popup.
type PopupAction struct {
	Type   string // "close_modal", "click_start", "scroll_play", "click_play", "new_game", "dismiss_sheet", "go_online"
	TapX   int
	TapY   int
	Swipe  bool
	SwipeX [2]int
	SwipeY [2]int
}

func isGreenButton(r, g, b int) bool {
	// Chess.com primary green button has g > 155, g-r > 45, g-b > 35
	return g > 155 && (g-r) > 45 && (g-b) > 35
}

func isBoardSquareColor(r, g, b int) bool {
	// Green
	dg := math.Sqrt(float64((r-118)*(r-118) + (g-150)*(g-150) + (b-86)*(b-86)))
	if dg < 35.0 {
		return true
	}
	// Cream
	dc := math.Sqrt(float64((r-238)*(r-238) + (g-238)*(g-238) + (b-210)*(b-210)))
	return dc < 35.0
}

// DetectContextualAction inspects the screen for any popups, dialogs, covered buttons, or post-game banners.
func DetectContextualAction(frame image.Image) *PopupAction {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	// 1. Post-Game Modal Overlay detection
	// When game ends, a dark modal banner appears at y ~ 480 px (y ~ 0.20), where background is dark (r,g,b < 50)
	bannerBg := frame.At(w/2, int(float64(h)*0.20))
	br, bg, bb, _ := bannerBg.RGBA()
	isDarkBanner := (br>>8) < 50 && (bg>>8) < 50 && (bb>>8) < 50

	if isDarkBanner {
		// End-game action button at y ~ 0.60
		reportBtnX := int(float64(w) * 0.50)
		for _, reportY := range []int{int(float64(h) * 0.446), int(float64(h) * 0.61)} {
			cReport := frame.At(reportBtnX, reportY)
			rr, rg, rb, _ := cReport.RGBA()
			if isGreenButton(int(rr>>8), int(rg>>8), int(rb>>8)) {
				// Navigate to home menu via logo tap
				return &PopupAction{
					Type: "go_home",
					TapX: int(float64(w) * 0.245),
					TapY: int(float64(h) * 0.079),
				}
			}
		}

		// Close button 'X' on end-game modal
		for _, cy := range []float64{0.190, 0.247} {
			egCloseX := int(float64(w) * 0.934)
			egCloseY := int(float64(h) * cy)
			cEgClose := frame.At(egCloseX, egCloseY)
			ecr, ecg, ecb, _ := cEgClose.RGBA()
			ecr8, ecg8, ecb8 := int(ecr>>8), int(ecg>>8), int(ecb>>8)
			if ecr8 > 110 && ecg8 > 110 && ecb8 > 110 {
				return &PopupAction{
					Type: "go_home",
					TapX: int(float64(w) * 0.245),
					TapY: int(float64(h) * 0.079),
				}
			}
		}
	}

	// 1b. Main Menu detection: primary action button
	// On Chess.com home menu, action button is located between y ~ 0.25 and 0.45 (y ~ 670..950 px)
	for _, testFrac := range []float64{0.30, 0.38, 0.35, 0.32} {
		testY := int(float64(h) * testFrac)
		cBtn := frame.At(w/2, testY)
		br, bg, bb, _ := cBtn.RGBA()
		if isGreenButton(int(br>>8), int(bg>>8), int(bb>>8)) {
			return &PopupAction{
				Type: "click_play_online",
				TapX: w / 2,
				TapY: testY,
			}
		}
	}

	// 2. Play button check (online lobby or bottom screen)
	// Located between y ~ 0.55 and 0.95 (e.g. y ~ 1447 px).
	var foundGreenY int
	foundGreenCount := 0
	for yFrac := 0.55; yFrac <= 0.95; yFrac += 0.01 {
		sy := int(float64(h) * yFrac)
		cs := frame.At(w/2, sy)
		sr, sg, sb, _ := cs.RGBA()
		if isGreenButton(int(sr>>8), int(sg>>8), int(sb>>8)) {
			foundGreenCount++
			if foundGreenY == 0 {
				foundGreenY = sy
			}
		}
	}

	if foundGreenCount >= 3 {
		// Button located below y=0.82 requires scroll up
		if float64(foundGreenY) > float64(h)*0.82 {
			return &PopupAction{
				Type:   "scroll_play",
				Swipe:  true,
				SwipeX: [2]int{w / 2, w / 2},
				SwipeY: [2]int{int(float64(h) * 0.78), int(float64(h) * 0.40)},
			}
		} else {
			return &PopupAction{
				Type: "click_play",
				TapX: w / 2,
				TapY: foundGreenY + 30,
			}
		}
	}

	// 3. Occlusion sheet detection over board
	topBoardCheckY := int(float64(h) * 0.10)
	midBoardCheckY := int(float64(h) * 0.42)
	cTop := frame.At(w/2, topBoardCheckY)
	tr, tg, tb, _ := cTop.RGBA()
	isTopBoard := isBoardSquareColor(int(tr>>8), int(tg>>8), int(tb>>8))

	cMid := frame.At(w/2, midBoardCheckY)
	mr, mg, mb, _ := cMid.RGBA()
	mr8, mg8, mb8 := int(mr>>8), int(mg>>8), int(mb>>8)

	if isTopBoard && (mr8 < 60 && mg8 < 60 && mb8 < 60) {
		return &PopupAction{
			Type:   "dismiss_sheet",
			Swipe:  true,
			SwipeX: [2]int{w / 2, w / 2},
			SwipeY: [2]int{int(float64(h) * 0.42), int(float64(h) * 0.75)},
		}
	}

	// 4. Modal Dialog Start button
	modalBtnX := int(float64(w) * 0.50)
	modalBtnY := int(float64(h) * 0.695)
	cModalBtn := frame.At(modalBtnX, modalBtnY)
	mbr, mbg, mbb, _ := cModalBtn.RGBA()
	if isGreenButton(int(mbr>>8), int(mbg>>8), int(mbb>>8)) {
		return &PopupAction{
			Type: "click_start",
			TapX: modalBtnX,
			TapY: modalBtnY,
		}
	}

	// 4. Modal "X" close button (standard modal)
	closeX := int(float64(w) * 0.922)
	closeY := int(float64(h) * 0.218)
	cClose := frame.At(closeX, closeY)
	cr, cg, cb, _ := cClose.RGBA()
	cr8, cg8, cb8 := int(cr>>8), int(cg>>8), int(cb>>8)
	if cr8 > 180 && cg8 > 180 && cb8 > 180 {
		bgX := int(float64(w) * 0.05)
		bgY := int(float64(h) * 0.20)
		cBg := frame.At(bgX, bgY)
		bgr, bgg, bgb, _ := cBg.RGBA()
		if (bgr>>8) < 50 && (bgg>>8) < 50 && (bgb>>8) < 50 {
			return &PopupAction{
				Type: "close_modal",
				TapX: closeX,
				TapY: closeY,
			}
		}
	}

	// 6. Check if we are in match with board shifted down:
	// Only run this if we are in an active match with top clock/player bar visible (y ~ 0.12 has player name or clock)
	clockCheckY := int(float64(h) * 0.136)
	cClock := frame.At(int(float64(w)*0.88), clockCheckY)
	crClock, cgClock, cbClock, _ := cClock.RGBA()
	clockLum := (int(crClock>>8) + int(cgClock>>8) + int(cbClock>>8)) / 3
	hasPlayerBar := clockLum > 80 // clock background is dark gray (~100) or white (~255)

	if hasPlayerBar {
		boardNormY := int(float64(h) * 0.16875) // normal board start (y ~ 405)
		cCheck := frame.At(w/2, boardNormY)
		cCheckR, cCheckG, cCheckB, _ := cCheck.RGBA()
		if !isBoardSquareColor(int(cCheckR>>8), int(cCheckG>>8), int(cCheckB>>8)) {
			// Check if board is visible lower down (e.g. y ~ 600..800)
			for testY := int(float64(h) * 0.25); testY < int(float64(h) * 0.45); testY += 50 {
				cT := frame.At(w/2, testY)
				tr, tg, tb, _ := cT.RGBA()
				if isBoardSquareColor(int(tr>>8), int(tg>>8), int(tb>>8)) {
					// Board is shifted down! Swipe up from mid to top to align board!
					return &PopupAction{
						Type:   "scroll_play",
						Swipe:  true,
						SwipeX: [2]int{w / 2, w / 2},
						SwipeY: [2]int{int(float64(h) * 0.60), int(float64(h) * 0.30)},
					}
				}
			}
		}
	}

	return nil
}

// DetectActiveBoardY scans the frame vertically for the 1064px square board.
func DetectActiveBoardY(frame image.Image) (int, bool) {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	// Candidates for chess.com mobile web:
	// Mode A: Uncovered match position (y = 0.16875 * h ~ 405)
	// Mode B: Top position (y = 0.06416 * h ~ 154)
	candidates := []int{
		int(float64(h) * 0.16875),
		int(float64(h) * 0.06416),
	}

	for _, yStart := range candidates {
		matches := 0
		cell := float64(w) * 0.985 / 8.0
		xStart := int(float64(w) * 0.0074)

		for r := 2; r <= 5; r++ {
			for c := 0; c < 8; c++ {
				sx := xStart + int((float64(c)+0.5)*cell)
				sy := yStart + int((float64(r)+0.5)*cell)
				if sy >= h {
					continue
				}
				px := frame.At(sx, sy)
				pr, pg, pb, _ := px.RGBA()
				pr8, pg8, pb8 := int(pr>>8), int(pg>>8), int(pb>>8)

				if isBoardSquareColor(pr8, pg8, pb8) {
					matches++
				}
			}
		}
		if matches >= 16 {
			return yStart, true
		}
	}

	return int(float64(h) * 0.16875), false
}

// DetectActiveClockTurn checks the header clocks to identify whose turn it is on Chess.com.
func DetectActiveClockTurn(frame image.Image, myTeam string) string {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	// Opponent clock [877,294][1012,358]: yFrac ~ 0.136, xFrac ~ 0.88
	// Player clock [877,1515][1012,1579]: yFrac ~ 0.645, xFrac ~ 0.88
	topClockY := int(float64(h) * 0.136)
	botClockY := int(float64(h) * 0.645)
	clockX := int(float64(w) * 0.88)

	cTop := frame.At(clockX, topClockY)
	cBot := frame.At(clockX, botClockY)

	tr, tg, tb, _ := cTop.RGBA()
	br, bg, bb, _ := cBot.RGBA()

	lumTop := (int(tr>>8) + int(tg>>8) + int(tb>>8)) / 3
	lumBot := (int(br>>8) + int(bg>>8) + int(bb>>8)) / 3

	otherTeam := "Black"
	if myTeam == "Black" {
		otherTeam = "White"
	}
	if myTeam == "" {
		myTeam = "White"
		otherTeam = "Black"
	}

	// Bottom clock belongs to player (myTeam), top clock belongs to opponent (otherTeam)
	if lumBot > 150 && lumTop < 120 {
		return myTeam
	}
	if lumTop > 150 && lumBot < 120 {
		return otherTeam
	}
	return ""
}

// GetPromotionQueenTap calculates tap coordinate for Queen promotion in Chess.com dialog.
func GetPromotionQueenTap(screenW, screenH int, toC int) (int, int) {
	cellW := float64(screenW) * 0.985 / 8.0
	x := int(float64(screenW)*0.0074 + (float64(toC)+0.5)*cellW)
	y := int(float64(screenH) * 0.20)
	return x, y
}
