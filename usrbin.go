package usrbin

// Option is a functional option for configuring the client
type Option func(*SDK) error

// UsingLogger can be set if you want to get log output from this sdk
// By default no log output is emitted.
func UsingLogger(logger Logger) Option {
	return func(sdk *SDK) error {
		sdk.logger = logger
		return nil
	}
}

// Using GitHubUpdateChecker will cause the repo passed in
// to be the source of truth when checking for new updates
func UsingGitHubUpdateChecker(repo string) Option {
	return func(sdk *SDK) error {
		sdk.updateChecker = NewGitHubUpdateChecker(repo)
		return nil
	}
}

// UsingHomebrewFormula will cause the formula passed in
// to be used when checking if this CLI was installed
// using homebrew
func UsingHomebrewFormula(formula string) Option {
	return func(sdk *SDK) error {
		sdk.externalPackageManagers = append(sdk.externalPackageManagers, NewHomebrewExternalPackageManager(formula))
		return nil
	}
}

func New(version string, opts ...Option) (*SDK, error) {
	sdk := SDK{
		version: version,
	}

	if err := sdk.parseOptions(opts); err != nil {
		return nil, err
	}

	return &sdk, nil
}

func (s *SDK) parseOptions(opts []Option) error {
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return err
		}
	}

	return nil
}
