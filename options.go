package godiff

type option interface {
	apply(*opts)
}

type opts struct {
	ignoreFields map[string]struct{}
}

// Option for ignoring fields
type IgnoreFields []string

func (i IgnoreFields) apply(opts *opts) {
	if opts.ignoreFields == nil {
		opts.ignoreFields = map[string]struct{}{}
	}

	for _, v := range i {
		opts.ignoreFields[v] = struct{}{}
	}
}

func IgnoreStructFields(name ...string) IgnoreFields {
	return IgnoreFields(name)
}
