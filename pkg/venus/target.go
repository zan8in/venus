package venus

import (
	"bufio"
	"io"
	"os"

	"github.com/remeh/sizedwaitgroup"
)

var TargetTempFile = "venus-target-temp-*"

func (r *Runner) PreprocessTarget() error {
	var (
		err error
	)

	targetTemp, err := os.CreateTemp("", TargetTempFile)
	if err != nil {
		return err
	}
	defer targetTemp.Close()

	if len(r.options.TargetFile) > 0 {
		f, err := os.Open(r.options.TargetFile)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(targetTemp, f); err != nil {
			return err
		}
	}

	f, err := os.Open(targetTemp.Name())
	if err != nil {
		return err
	}
	defer f.Close()

	wg := sizedwaitgroup.New(100)
	s := bufio.NewScanner(f)
	for s.Scan() {
		wg.Add()
		go func(target string) {
			defer wg.Done()
			r.options.Target.Set(target)
		}(s.Text())
	}
	wg.Wait()

	return err
}
