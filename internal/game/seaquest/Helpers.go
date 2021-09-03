package seaquest

import "github.com/samuelfneumann/goatar/internal/game"

// submarine implements a submarine in the SeaQuest game
type submarine struct {
	*swimmer
	shotTimer int // Can only shoot once this reaches 0
}

// newSubmarine returns a new submarine
func newSubmarine(x, y int, right bool, moveTimer, shotTimer int) *submarine {
	swimmer := newSwimmer(x, y, right, moveTimer)

	sub := &submarine{
		swimmer:   swimmer,
		shotTimer: shotTimer,
	}

	return sub
}

// canShoot returns whether the submarine is allowed to shoot yet or not
func (s *submarine) canShoot() bool {
	return s.shotTimer == 0
}

// setShotTimer sets the timer for the submarine to shoot again
func (s *submarine) setShotTimer(val int) {
	s.shotTimer = val
}

// decrementShotTimer decrements the shot timer
func (s *submarine) decrementShotTimer() {
	if s.shotTimer > 0 {
		s.shotTimer--
	}
}

// swimmer implements functionality for any entity in the SeaQuest game
// that can swim or move underwater
type swimmer struct {
	xPos          int
	yPos          int
	moveDirection int
	moveTimer     int // Can only move once this reaches 0
}

// newSwimmer returns a new swimmer
func newSwimmer(x, y int, right bool, moveTimer int) *swimmer {
	var direction int
	if right {
		direction = 1
	} else {
		direction = -1
	}

	return &swimmer{
		xPos:          x,
		yPos:          y,
		moveDirection: direction,
		moveTimer:     moveTimer,
	}
}

// newBullet returns a new bullet, which is a swimmer with no move timer
func newBullet(x, y int, right bool) *swimmer {
	return newSwimmer(x, y, right, 0)
}

// direction returns the direction of movement of the swimmer. +1
// indicates movement right, and -1 indicates movement left.
func (s *swimmer) direction() int {
	return s.moveDirection
}

// setDirection sets the direction of movement. The parameter right
// denotes whether or not the direction of movement should be to the
// right.
func (s *swimmer) setDirection(right bool) {
	if right {
		s.moveDirection = 1
	} else {
		s.moveDirection = -1
	}
}

// y returns the y position of the swimmer
func (s *swimmer) y() int {
	return s.yPos
}

// setY sets the y position of the swimmer
func (s *swimmer) setY(pos int) {
	s.yPos = pos
}

// x returns the x position of the swimmer
func (s *swimmer) x() int {
	return s.xPos
}

// setX sets the x position of the swimmer
func (s *swimmer) setX(pos int) {
	s.xPos = pos
}

// orientedRight returns whether the swimmer is oriented/moving to the
// right or not
func (s *swimmer) orientedRight() bool {
	return s.direction() == 1
}

// move moves the swimmer in the direction of movement
func (s *swimmer) move() {
	s.xPos += s.direction()
}

// canMove returns whether the swimmer can move or not
func (s *swimmer) canMove() bool {
	return s.moveTimer == 0
}

// decrementMoveTimer decrements the move timer of the swimmer
func (s *swimmer) decrementMoveTimer() {
	s.moveTimer--
}

// setMoveTimer sets the move timer of the swimmer
func (s *swimmer) setMoveTimer(val int) {
	s.moveTimer = val
}

// player implements the player in the SeaQuest game
type player struct {
	*submarine
	remainingOxygen int
	diverCount      int
}

// newPlayer returns a new player
func newPlayer(x, y int, right bool, moveTimer, shotTimer,
	oxygen int) *player {
	sub := newSubmarine(x, y, right, moveTimer, shotTimer)

	return &player{
		submarine:       sub,
		remainingOxygen: oxygen,
		diverCount:      0,
	}
}

// setDivers sets the number of divers held in the player's submarine
func (p *player) setDivers(num int) {
	p.diverCount = num
}

// divers returns the number of divers in the player's submarine
func (p *player) divers() int {
	return p.diverCount
}

// decrementDivers removes one diver from the player's submarine
func (p *player) decrementDivers() {
	p.setDivers(p.divers() - 1)
}

// incrementDivers adds one diver to the player's submarine
func (p *player) incrementDivers() {
	p.setDivers(p.divers() + 1)
}

// oxygen returns the current oxygen level in the player's submarine
func (p *player) oxygen() int {
	return p.remainingOxygen
}

// setOxygen sets the oxygen level in the player's submarine to level
func (p *player) setOxygen(level int) {
	p.remainingOxygen = level
}

// decrementOxygen removes one unit of oxygen from the player's
// submarine
func (p *player) decrementOxygen() {
	p.setOxygen(p.oxygen() - 1)
}

// moveLeft moves the player left
func (p *player) moveLeft() {
	p.setX(game.MaxInt(0, p.x()-1))
	p.setDirection(false)
}

// moveRight moves the player right
func (p *player) moveRight() {
	p.setX(game.MinInt(cols-1, p.x()+1))
	p.setDirection(true)
}

// moveDown moves the player down
func (p *player) moveDown() {
	p.setY(game.MinInt(rows-2, p.y()+1))
}

// moveUp moves the player up
func (p *player) moveUp() {
	p.setY(game.MaxInt(0, p.y()-1))
}
