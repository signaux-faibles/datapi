package main

import "fmt"

type htree struct {
	one      *htree
	zero     *htree
	two      *htree
	three    *htree
	four     *htree
	five     *htree
	six      *htree
	seven    *htree
	height   *htree
	nine     *htree
	ten      *htree
	eleven   *htree
	twelve   *htree
	thirteen *htree
	fourteen *htree
	fifteen  *htree
}

func (t *htree) insert(hash []byte) *htree {
	current := t
	bs := fmt.Sprintf("%x", hash)
	for _, i := range bs {
		switch i {
		case '0':
			if current.zero == nil {
				current.zero = &htree{}
			}
			current = current.zero
		case '1':
			if current.one == nil {
				current.one = &htree{}
			}
			current = current.one
		case '2':
			if current.two == nil {
				current.two = &htree{}
			}
			current = current.two
		case '3':
			if current.three == nil {
				current.three = &htree{}
			}
			current = current.three
		case '4':
			if current.four == nil {
				current.four = &htree{}
			}
			current = current.four
		case '5':
			if current.five == nil {
				current.five = &htree{}
			}
			current = current.five
		case '6':
			if current.six == nil {
				current.six = &htree{}
			}
			current = current.six
		case '7':
			if current.seven == nil {
				current.seven = &htree{}
			}
			current = current.seven
		case '8':
			if current.height == nil {
				current.height = &htree{}
			}
			current = current.height
		case '9':
			if current.nine == nil {
				current.nine = &htree{}
			}
			current = current.nine
		case 'a':
			if current.ten == nil {
				current.ten = &htree{}
			}
			current = current.ten
		case 'b':
			if current.eleven == nil {
				current.eleven = &htree{}
			}
			current = current.eleven
		case 'c':
			if current.twelve == nil {
				current.twelve = &htree{}
			}
			current = current.twelve
		case 'd':
			if current.thirteen == nil {
				current.thirteen = &htree{}
			}
			current = current.thirteen
		case 'e':
			if current.fourteen == nil {
				current.fourteen = &htree{}
			}
			current = current.fourteen
		case 'f':
			if current.fifteen == nil {
				current.fifteen = &htree{}
			}
			current = current.fifteen
		}
	}
	return current
}

func (t *htree) read(b []byte) *htree {
	current := t
	bs := fmt.Sprintf("%x", b)
	for _, i := range bs {
		switch i {
		case '0':
			if current.zero == nil {
				return nil
			}
			current = current.zero
		case '1':
			if current.one == nil {
				return nil
			}
			current = current.one
		case '2':
			if current.two == nil {
				return nil
			}
			current = current.two
		case '3':
			if current.three == nil {
				return nil
			}
			current = current.three
		case '4':
			if current.four == nil {
				return nil
			}
			current = current.four
		case '5':
			if current.five == nil {
				return nil
			}
			current = current.five
		case '6':
			if current.six == nil {
				return nil
			}
			current = current.six
		case '7':
			if current.seven == nil {
				return nil
			}
			current = current.seven
		case '8':
			if current.height == nil {
				return nil
			}
			current = current.height
		case '9':
			if current.nine == nil {
				return nil
			}
			current = current.nine
		case 'a':
			if current.ten == nil {
				return nil
			}
			current = current.ten
		case 'b':
			if current.eleven == nil {
				return nil
			}
			current = current.eleven
		case 'c':
			if current.twelve == nil {
				return nil
			}
			current = current.twelve
		case 'd':
			if current.thirteen == nil {
				return nil
			}
			current = current.thirteen
		case 'e':
			if current.fourteen == nil {
				return nil
			}
			current = current.fourteen
		case 'f':
			if current.fifteen == nil {
				return nil
			}
			current = current.fifteen
		}
	}
	return current
}

func (t *htree) contains(hash []byte) bool {
	current := t.read(hash)
	if current == nil {
		return false
	}
	return true
}
