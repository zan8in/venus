package subdomain

var (
	FakeDomain    = "venus-test-dns-a.venus.com"
	FakeSubdomain = "venus-test-dns-a"
	BaiduDNS      = "public-dns-a.baidu.com"
	BaiduIP       = "180.76.76.76"

	DefaultResolvers = []string{
		"114.114.114.114:53",
		"223.5.5.5:53",       // ali dns
		"180.76.76.76:53",    // baidu dns
		"119.29.29.29:53",    // tencent dns
		"182.254.118.118:53", // tencent dns
		"1.1.1.1:53",         // Cloudflare primary
		"1.0.0.1:53",         // Cloudflare secondary
		"8.8.8.8:53",         // Google primary
		"8.8.4.4:53",         // Google secondary
		"9.9.9.9:53",         // Quad9 Primary
		"9.9.9.10:53",        // Quad9 Secondary
		"77.88.8.8:53",       // Yandex Primary
		"77.88.8.1:53",       // Yandex Secondary
		"208.67.222.222:53",  // OpenDNS Primary
		"208.67.220.220:53",  // OpenDNS Secondary
	}

	DefaultReadDictRateTime = 100

	SubNameTempFile = "venus-subname-temp-*"
)
