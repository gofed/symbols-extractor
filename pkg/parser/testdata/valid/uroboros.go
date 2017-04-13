package uroboros

type Snake struct {
	Head Head
}

type Head struct {
	Tail Tail
}

type Tail struct {
	Snake *Snake
}

func CreateSnake() Snake {
	ka := Snake{}
	ka.Head.Tail.Snake = &ka
	return ka
}
