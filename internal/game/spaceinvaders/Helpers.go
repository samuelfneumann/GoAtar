package spaceinvaders

import "github.com/samuelfneumann/goatar/internal/game"

type player struct {
	position  int
	shotTimer int
}

func newPlayer(position, shotTimer int) *player {
	return &player{position, shotTimer}
}

func (p *player) x() int {
	return p.position
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

func (p *player) setX(pos int) {
	p.position = pos
}

// moveLeft moves the player left
func (p *player) moveLeft() {
	p.setX(game.MaxInt(0, p.x()-1))
}

// moveRight moves the player right
func (p *player) moveRight() {
	p.setX(game.MinInt(cols-1, p.x()+1))
}
