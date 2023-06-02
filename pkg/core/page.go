package core

type Page struct {
	Size   int
	Number int
}

func (p Page) From() int {
	return p.Number * p.Size
}
