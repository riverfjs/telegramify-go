package telegramify

// ConvertOptions holds options for markdown conversion.
type ConvertOptions struct {
	LatexEscape bool
	Config      *RenderConfig
}

// Option is a function that configures ConvertOptions.
type Option func(*ConvertOptions)

// WithLatexEscape sets whether to convert LaTeX to Unicode.
func WithLatexEscape(enable bool) Option {
	return func(opts *ConvertOptions) {
		opts.LatexEscape = enable
	}
}

// WithConfig sets a custom RenderConfig.
func WithConfig(config *RenderConfig) Option {
	return func(opts *ConvertOptions) {
		opts.Config = config
	}
}

// defaultConvertOptions returns the default conversion options.
func defaultConvertOptions() *ConvertOptions {
	return &ConvertOptions{
		LatexEscape: true,
		Config:      DefaultConfig(),
	}
}

// applyOptions applies the given options to the default options.
func applyOptions(opts ...Option) *ConvertOptions {
	options := defaultConvertOptions()
	for _, opt := range opts {
		opt(options)
	}
	return options
}

