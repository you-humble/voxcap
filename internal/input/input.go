package input

// KeyReader abstracts keyboard input.
type KeyReader interface {
	ReadKey() (rune, error)
	Close()
}
