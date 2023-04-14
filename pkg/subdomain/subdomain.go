package subdomain

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zan8in/gologger"
	"github.com/zan8in/pavo"
	sliceutil "github.com/zan8in/pins/slice"
	"github.com/zan8in/venus/pkg/util/setutil"
)

type SubDomain struct {
	ctx              context.Context
	timeout          time.Duration
	resolver         *net.Resolver
	WildcardIps      setutil.Set[netip.Addr]
	BlacklistedIps   setutil.Set[string]
	ResultSubdomains sliceutil.SafeSlice
	isWildcard       bool
	dnsList          []string

	dictTempName string
	DictChan     chan string

	rateLimit int
}

func newCustomDialer(dns string) func(ctx context.Context, network, address string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{
			Timeout:   time.Second,
			KeepAlive: time.Second,
		}
		return d.DialContext(ctx, "udp", dns)
	}
}

func NewSuDomain(timeout, ratelimit int) (*SubDomain, error) {

	subdomain := &SubDomain{
		ctx:              context.Background(),
		timeout:          time.Duration(timeout) * time.Second,
		resolver:         &net.Resolver{},
		isWildcard:       false,
		WildcardIps:      setutil.NewSet[netip.Addr](),
		BlacklistedIps:   setutil.Set[string]{},
		ResultSubdomains: sliceutil.SafeSlice{},
		DictChan:         make(chan string),
		rateLimit:        ratelimit,
	}

	if err := subdomain.currentDNSServers(); err != nil {
		return nil, err
	}

	if err := subdomain.PreprocessDict(); err != nil {
		return nil, err
	}

	return subdomain, nil
}

func (s *SubDomain) DetectWildcardParse(domain string) error {
	rand.Seed(time.Now().UnixNano())

	count := 0
	for i := 0; i < 20; i++ {
		subdomain := fmt.Sprintf("%d.%d.%d.%d.%s", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256), domain)
		ips, err := net.LookupIP(subdomain)
		if err == nil && len(ips) > 0 {
			ip := ips[0]
			for _, otherIP := range ips[1:] {
				if ip.Equal(otherIP) {
					count++
				}
			}
			if count > 10 {
				na, _ := netip.ParseAddr(ip.String())
				s.WildcardIps.Add(na)
				return fmt.Errorf("this domain name `%s` resolved to %s, and the program will automatically scan the domain name resolved to this IP (%s)", domain, ip.String(), ip.String())
			}
		}
	}

	return nil
}

func (s *SubDomain) detectWildcard(ctx context.Context, domain string) bool {
	var num int32
	wg := &sync.WaitGroup{}
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			domain = fmt.Sprintf("%s-%d.%s", FakeSubdomain, i, domain)
			if _, err := s.dnsLookup(ctx, domain); err == nil {
				atomic.AddInt32(&num, 1)
			}
		}(i)

	}
	wg.Wait()

	return num >= 3
}

func (s *SubDomain) currentDNSServers() (err error) {
	wg := &sync.WaitGroup{}
	for _, dns := range DefaultResolvers {
		wg.Add(1)
		go func(dns string) {
			defer wg.Done()

			if d, err := s.detectDNSServer(dns); err == nil && len(d) > 0 {
				s.dnsList = append(s.dnsList, d)
			}
		}(dns)

	}
	wg.Wait()

	if len(s.dnsList) == 0 {
		return fmt.Errorf("no DNS servers")
	}
	return nil
}

// Check if the DNS server can give the correct response
//
// return current dns server
//
// Use the correct domain name to test whether a DNS can send the correct response
// Then use a wrong domain name to test DNS, if it can respond correctly, it means that the DNS is not reliable
func (s *SubDomain) detectDNSServer(dns string) (string, error) {
	var resolver = &net.Resolver{
		PreferGo: true,
		Dial:     newCustomDialer(dns),
	}
	s.resolver = resolver

	ips, err := s.dnsLookup(s.ctx, BaiduDNS)
	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			return "", err
		}
		return "", err
	}
	for _, ip := range ips {
		if ip.String() == BaiduIP {

			_, err := s.dnsLookup(s.ctx, FakeDomain)
			if err != nil {
				return dns, nil
			}

		}
	}

	return "", fmt.Errorf("fake dns server")
}

func (s *SubDomain) DnsLookupRandomResolver(domain string) ([]netip.Addr, error) {
	s.resolver = &net.Resolver{
		PreferGo: true,
		Dial:     newCustomDialer(s.randomDNS()),
	}
	ctx2, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.resolver.LookupNetIP(ctx2, "ip", domain)
}

func (s *SubDomain) dnsLookup(ctx context.Context, domain string) ([]netip.Addr, error) {
	ctx2, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.resolver.LookupNetIP(ctx2, "ip", domain)
}

func (s *SubDomain) dnsLookupCname(ctx context.Context, domain string) (string, error) {
	ctx2, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	time.Sleep(time.Second)
	return s.resolver.LookupCNAME(ctx2, domain)
}

func (s *SubDomain) randomDNS() string {
	if len(s.dnsList) == 0 {
		return DefaultResolvers[7]
	}
	return s.dnsList[rand.New(rand.NewSource(time.Now().Unix())).Intn(len(s.dnsList))]
}

func (s *SubDomain) PavoSubdomain(domain string) ([]string, error) {
	r, err := pavo.QuerySubDomain(domain, 1000)
	if err != nil {
		err := fmt.Errorf("%s, please edit file `%s`", err.Error(), s.PavoConfigName())
		gologger.Info().Msg(err.Error())
		return nil, err
	}
	return r, nil
}

func (s *SubDomain) PavoConfigName() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configDir := filepath.Join(homeDir, ".config", "pavo")

	return filepath.Join(configDir, "pavo.yml")
}
