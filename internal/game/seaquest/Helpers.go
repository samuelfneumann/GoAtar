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

func (s *swimmer) setDirection(right bool) {
	if right {
		s.moveDirection = 1
	} else {
		s.moveDirection = -1
	}
}

func (s *swimmer) y() int {
	return s.yPos
}

func (s *swimmer) setY(pos int) {
	s.yPos = pos
}

func (s *swimmer) x() int {
	return s.xPos
}

func (s *swimmer) setX(pos int) {
	s.xPos = pos
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

type player struct {
	*submarine
	remainingOxygen int
	diverCount      int
}

func newPlayer(x, y int, right bool, moveTimer, shotTimer,
	oxygen int) *player {
	sub := newSubmarine(x, y, right, moveTimer, shotTimer)

	return &player{
		submarine:       sub,
		remainingOxygen: oxygen,
		diverCount:      0,
	}
}

func (p *player) setDivers(num int) {
	p.diverCount = num
}

func (p *player) divers() int {
	return p.diverCount
}

func (p *player) decrementDivers() {
	p.setDivers(p.divers() - 1)
}

func (p *player) incrementDivers() {
	p.setDivers(p.divers() + 1)
}

func (p *player) oxygen() int {
	return p.remainingOxygen
}

func (p *player) setOxygen(level int) {
	p.remainingOxygen = level
}

func (p *player) decrementOxygen() {
	p.setOxygen(p.oxygen() - 1)
}

func (p *player) incrementOxygen() {
	p.setOxygen(p.oxygen() + 1)
}

func (p *player) moveLeft() {
	p.setX(maxInt(0, p.x()-1))
	p.setDirection(false)
}

func (p *player) moveRight() {
	p.setX(minInt(rows-1, p.x()+1))
	p.setDirection(true)
}

func (p *player) moveDown() {
	p.setY(minInt(rows-2, p.y()+1))
}

func (p *player) moveUp() {
	p.setY(maxInt(0, p.y()-1))
}
