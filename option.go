package laurel

type Option func(opts *Options)

func loadOptions(options ...Option) *Options {
	opts := new(Options)
	for _, option := range options {
		option(opts)
	}
	return opts
}

// Options contains all options which will be applied
type Options struct {
	IsTestCmd bool
	InputCmd  <-chan string
	ResMsg    chan string
}

// WithOptions accepts the whole options config.
func WithOptions(options Options) Option {
	return func(opts *Options) {
		*opts = options
	}
}

func WithIsTestCmd(IsTestCmd bool) Option {
	return func(opts *Options) {
		opts.IsTestCmd = IsTestCmd
	}
}

func WithInputCmd(InputCmd <-chan string) Option {
	if InputCmd == nil {
		InputCmd = make(<-chan string)
	}
	return func(opts *Options) {
		opts.InputCmd = InputCmd
	}
}

func WithResMsg(ResMsg chan string) Option {
	if ResMsg == nil {
		ResMsg = make(chan string)
	}
	return func(opts *Options) {
		opts.ResMsg = ResMsg
	}
}
