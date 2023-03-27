package venus

import (
	"fmt"
	"time"

	"github.com/remeh/sizedwaitgroup"
	"github.com/zan8in/gologger"
	"github.com/zan8in/venus/pkg/result"
	"github.com/zan8in/venus/pkg/subdomain"
	"github.com/zan8in/venus/pkg/util/urlutil"
)

type (
	Runner struct {
		options *Options

		subdomain *subdomain.SubDomain

		ticker    *time.Ticker
		wgscan    sizedwaitgroup.SizedWaitGroup
		rateLimit int

		result *result.Result
	}
)

func NewRunner(options *Options) (*Runner, error) {
	runner := &Runner{
		options:   options,
		result:    result.NewResult(),
		rateLimit: options.RateLimit,
	}

	s, err := subdomain.NewSuDomain(
		runner.options.Timeout,
		runner.options.RateLimit,
	)
	if err != nil {
		return runner, err
	}

	runner.subdomain = s

	runner.wgscan = sizedwaitgroup.New(runner.rateLimit)
	runner.ticker = time.NewTicker(time.Second / time.Duration(runner.rateLimit))

	return runner, nil
}

func (runner *Runner) Run() error {
	go runner.subdomain.DictList()

	if err := runner.Blaster(runner.options.Target); err != nil {
		return err
	}

	return nil
}

func (r *Runner) Blaster(domains []string) error {
	for _, domain := range domains {
		if len(urlutil.TopDomain(domain)) == 0 {
			r.subdomain.BlacklistedIps.Add(domain)
			gologger.Info().Msgf("%s is not a valid domain", domain)
		}

		if err := r.subdomain.DetectWildcardParse(domain); err != nil {
			gologger.Info().Msg(err.Error())
		}
	}

	if r.subdomain.DictChan == nil {
		return fmt.Errorf("no dict channel")
	}

	for d := range r.subdomain.DictChan {
		for _, domain := range domains {
			if r.subdomain.BlacklistedIps.Contains(domain) {
				continue
			}
			r.wgscan.Add()
			go func(name, domain string) {
				defer r.wgscan.Done()
				<-r.ticker.C

				sub := fmt.Sprintf("%s.%s", name, domain)
				if ips, err := r.subdomain.DnsLookupRandomResolver(sub); err == nil {
					if !r.subdomain.WildcardIps.ContainsAny(ips) {
						gologger.Info().Msg(sub)
					}
				}

			}(d, domain)
		}
	}
	r.wgscan.Wait()

	return nil
}