package spaceinvaders

import "github.com/samuelfneumann/goatar/internal/game"

// player implements a player in the game SpaceInvaders
type player struct {
	position  int
	shotTimer int
}

// newPlayer returns a new player
func newPlayer(position, shotTimer int) *player {
	return &player{position, shotTimer}
}

// x returns the x position of the player
func (p *player) x() int {
	return p.position
}

// setX sets the x position of the player
func (p *player) setX(pos int) {
	p.position = pos
}

// canShoot returns whether or not the player can shoot
func (p *player) canShoot() bool {
	return p.shotTimer <= 0
}

// setShotTimer sets the time until the player can shoot again
func (p *player) setShotTimer(time int) {
	p.shotTimer = time
}

// decrementShotTimer decrements the time until the player can shoot
// again
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
