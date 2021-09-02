package seaquest

type bullet struct {
	xPos      int
	yPos      int
	direction int
}

func newBullet(x, y int, right bool) *bullet {
	var direction int
	if right {
		direction = 1
	} else {
		direction = -1
	}
	return &bullet{x, y, direction}
}

func (b *bullet) move() {
	b.xPos += b.direction
}

func (b *bullet) y() int {
	return b.yPos
}

func (b *bullet) x() int {
	return b.xPos
}

type submarine struct {
	*swimmer
	shotTimer int
}

func newSubmarine(x, y int, right bool, moveTimer, shotTimer int) *submarine {
	swimmer := newSwimmer(x, y, right, moveTimer)

	sub := &submarine{
		swimmer:   swimmer,
		shotTimer: shotTimer,
	}

	return sub
}

func (s *submarine) canShoot() bool {
	return s.shotTimer == 0
}

func (s *submarine) setShotTimer(val int) {
	s.shotTimer = val
}

func (s *submarine) decrementShotTimer() {
	s.shotTimer--
}

type swimmer struct {
	xPos          int
	yPos          int
	moveDirection int
	moveTimer     int
}

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

func (s *swimmer) direction() int {
	return s.moveDirection
}

func (s *swimmer) y() int {
	return s.yPos
}

func (s *swimmer) x() int {
	return s.xPos
}

func (s *swimmer) orientedRight() bool {
	return s.direction() == 1
}

func (s *swimmer) move() {
	s.xPos += s.direction()
}

func (s *swimmer) canMove() bool {
	return s.moveTimer == 0
}

func (s *swimmer) decrementMoveTimer() {
	s.moveTimer--
}

func (s *swimmer) setMoveTimer(val int) {
	s.moveTimer = val
}
