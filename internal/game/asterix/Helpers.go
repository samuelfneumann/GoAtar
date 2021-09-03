package asterix

import "github.com/samuelfneumann/goatar/internal/game"

// player implements a player in the game Asterix
type player struct {
	xPos      int
	yPos      int
	moveTimer int // Player can move once this reaches 0
}

// newPlayer returns a new player
func newPlayer(x, y, moveTimer int) *player {
	return &player{x, y, moveTimer}
}

// y returns the y position of the player
func (p *player) y() int {
	return p.yPos
}

// setY sets the y position of the player
func (p *player) setY(y int) {
	p.yPos = y
}

// x returns the x position of the player
func (p *player) x() int {
	return p.xPos
}

// setX sets the x position of the player
func (p *player) setX(x int) {
	p.xPos = x
}

// canMove returns whether or not the player can move
func (p *player) canMove() bool {
	return p.moveTimer <= 0
}

// setMoveTimer sets the time before the player can move
func (p *player) setMoveTimer(time int) {
	p.moveTimer = time
}

// decrementMoveTimer decrements the move timer
func (p *player) decrementMoveTimer() {
	if p.moveTimer > 0 {
		p.moveTimer--
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

// Entity implements an entity in the Asterix game, which is either an
// enemy or a gold
type Entity struct {
	xPos          int
	yPos          int
	moveDirection int
	gold          bool
}

// newEntity returns a new entity
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

// move moves the entity in its movement direction
func (e *Entity) move() {
	e.xPos += e.moveDirection
}

// isGold returns whether the entity is gold or not
func (e *Entity) isGold() bool {
	return e.gold
}

// direction returns the direction of movement of the entity
func (e *Entity) direction() int {
	return e.moveDirection
}

// orientedRight returns whether the entity is moving to the right
func (e *Entity) orientedRight() bool {
	return e.direction() == 1
}

// x returns the x position of the entity
func (e *Entity) x() int {
	return e.xPos
}

// y returns the y position of the entity
func (e *Entity) y() int {
	return e.yPos
}
