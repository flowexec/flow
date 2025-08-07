package builder

type Option func(*OptionsValue)

type OptionsValue struct {
	WorkspaceName string
	WorkspacePath string
	NamespaceName string
	FlowFilePath  string
}

func WithWorkspaceName(name string) Option {
	return func(opts *OptionsValue) {
		opts.WorkspaceName = name
	}
}

func WithWorkspacePath(path string) Option {
	return func(opts *OptionsValue) {
		opts.WorkspacePath = path
	}
}

func WithNamespaceName(name string) Option {
	return func(opts *OptionsValue) {
		opts.NamespaceName = name
	}
}

func WithFlowFilePath(path string) Option {
	return func(opts *OptionsValue) {
		opts.FlowFilePath = path
	}
}

func NewOptionValues(opts ...Option) *OptionsValue {
	o := &OptionsValue{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}
