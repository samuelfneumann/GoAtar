package asterix

import "github.com/samuelfneumann/goatar/internal/game"

type player struct {
	xPos      int
	yPos      int
	moveTimer int
	shotTimer int
}

func newPlayer(x, y, shotTimer, moveTimer int) *player {
	return &player{x, y, shotTimer, moveTimer}
}

func (p *player) y() int {
	return p.yPos
}

func (p *player) setY(y int) {
	p.yPos = y
}

func (p *player) x() int {
	return p.xPos
}

func (p *player) setX(x int) {
	p.xPos = x
}

func (p *player) canMove() bool {
	return p.moveTimer == 0
}

func (p *player) setMoveTimer(time int) {
	p.moveTimer = time
}

func (p *player) decrementMoveTimer() {
	if p.moveTimer > 0 {
		p.moveTimer--
	}
}

func (p *player) canShoot() bool {
	return p.shotTimer <= 0
}

func (p *player) setShotTimer(time int) {
	p.shotTimer = time
}

func (p *player) decrementShotTimer() {
	if p.shotTimer > 0 {
		p.shotTimer--
	}
}

// moveLeft moves the player left
func (p *player) moveLeft() {
	p.setX(game.MaxInt(0, p.x()-1))
}

// moveRight moves the player right
func (p *player) moveRight() {
	p.setX(game.MinInt(cols-1, p.x()+1))
}

// moveUp moves the player up
func (p *player) moveUp() {
	p.setY(game.MaxInt(1, p.y()-1))
}

// moveDown moves the player down
func (p *player) moveDown() {
	p.setY(game.MinInt(rows-2, p.y()+1))
}

type Entity struct {
	xPos          int
	yPos          int
	moveDirection int
	gold          bool
}

func newEntity(x, y int, orientedRight, isGold bool) *Entity {
	direction := -1
	if orientedRight {
		direction = 1
	}

	return &Entity{
		xPos:          x,
		yPos:          y,
		moveDirection: direction,
		gold:          isGold,
	}
}

func (e *Entity) move() {
	e.xPos += e.moveDirection
}

func (e *Entity) isGold() bool {
	return e.gold
}

func (e *Entity) direction() int {
	return e.moveDirection
}

func (e *Entity) orientedRight() bool {
	return e.direction() == 1
}

func (e *Entity) x() int {
	return e.xPos
}

func (e *Entity) y() int {
	return e.yPos
}
