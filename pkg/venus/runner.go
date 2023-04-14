package venus

import (
	"fmt"
	"time"

	"github.com/remeh/sizedwaitgroup"
	"github.com/zan8in/gologger"
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
	}
)

func NewRunner(options *Options) (*Runner, error) {
	runner := &Runner{
		options:   options,
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
	defer runner.sendDone()

	go runner.subdomain.DictList()

	if err := runner.Blaster(runner.options.Target); err != nil {
		return err
	}

	return nil
}

func (r *Runner) send(domain, sub string) {
	ds := fmt.Sprintf("%s:%s", domain, sub)
	if r.subdomain.ResultSubdomains.Key(ds) == -1 || r.subdomain.ResultSubdomains.Len() == 0 {
		r.subdomain.ResultSubdomains.Append(ds)
		rst := map[string]string{}
		rst[domain] = sub
		r.options.OnResult(rst)
	}
}

func (r *Runner) sendDone() {
	rst := map[string]string{}
	rst["DONE"] = "DONE"
	r.options.OnResult(rst)
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

		if subs, err := r.subdomain.PavoSubdomain(domain); err == nil {
			for _, sub := range subs {
				r.send(domain, sub)
			}
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
						r.send(domain, sub)
					}
				}

			}(d, domain)
		}
	}
	r.wgscan.Wait()

	return nil
}
