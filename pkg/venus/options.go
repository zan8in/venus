package venus

import (
	"runtime"

	"github.com/pkg/errors"
	"github.com/zan8in/goflags"
	"github.com/zan8in/gologger"
)

type OnResultCallback func(map[string]string)

type (
	Options struct {
		Target goflags.StringSlice

		RateLimit int // RateLimit is the rate of port
		Timeout   int //

		OnResult OnResultCallback
	}
)

func ParseOptions() *Options {
	options := &Options{}

	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`Venus`)

	flagSet.CreateGroup("input", "Input",
		flagSet.StringSliceVarP(&options.Target, "t", "target", nil, "target to scan subdomains for (comma-separated)", goflags.NormalizedStringSliceOptions),
	)

	flagSet.CreateGroup("rate-limit", "Rate-limit",
		flagSet.IntVar(&options.RateLimit, "rate", DefaultRateLimit, "packets to send per second"),
	)

	flagSet.CreateGroup("optimization", "Optimization",
		flagSet.IntVar(&options.Timeout, "timeout", DefaultTimeout, "millisecond to wait before timing out"),
	)

	_ = flagSet.Parse()

	err := options.validateOptions()
	if err != nil {
		gologger.Fatal().Msgf("Program exiting: %s\n", err)
	}

	return options
}

var (
	errNoInputList = errors.New("no input list provided")
	errZeroValue   = errors.New("cannot be zero")
)

func (options *Options) validateOptions() (err error) {

	if options.Target == nil {
		return errNoInputList
	}

	if options.RateLimit <= 0 {
		return errors.Wrap(errZeroValue, "rate")
	} else if options.RateLimit == DefaultRateLimit {
		options.autoChangeRateLimit()
	}

	if options.Timeout <= 0 {
		return errors.Wrap(errZeroValue, "timeout")
	} else {
		options.Timeout = DefaultTimeout
	}

	options.OnResult = func(map[string]string) {}

	return err
}

func (options *Options) autoChangeRateLimit() {
	NumCPU := runtime.NumCPU()
	options.RateLimit = NumCPU * 50
}
