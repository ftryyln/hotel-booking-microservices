package query

// Options defines basic pagination/filter inputs.
type Options struct {
	Limit  int
	Offset int
}

// Normalize applies sensible defaults and guards negatives.
func (o Options) Normalize(defaultLimit int) Options {
	if defaultLimit <= 0 {
		defaultLimit = 50
	}
	if o.Limit <= 0 {
		o.Limit = defaultLimit
	}
	if o.Offset < 0 {
		o.Offset = 0
	}
	return o
}
