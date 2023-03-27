package result

import "sync"

type Result struct {
	sync.RWMutex
	domains map[string][]string
}

func NewResult() *Result {
	return &Result{
		domains: make(map[string][]string),
	}
}

func (r *Result) AddResult(domain, subdomain string) {
	r.RLock()
	defer r.RUnlock()

	r.domains[domain] = append(r.domains[domain], subdomain)
}

func (r *Result) HasResult() bool {
	r.RLock()
	defer r.Unlock()

	return len(r.domains) > 0
}

func (r *Result) GetDomainResults(domain string) chan string {
	r.Lock()

	out := make(chan string)

	go func() {
		defer close(out)
		defer r.Unlock()

		for _, domain := range r.domains[domain] {
			out <- domain
		}
	}()

	return out
}
