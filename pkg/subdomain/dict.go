package subdomain

import (
	"bufio"
	"embed"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/remeh/sizedwaitgroup"
)

//go:embed dict/*
var f embed.FS

func (s *SubDomain) PreprocessDict() (err error) {

	subnameTemp, err := os.CreateTemp("", SubNameTempFile)
	if err != nil {
		return err
	}
	defer subnameTemp.Close()

	if len(s.Dict) > 0 {
		f, err := os.Open(s.Dict)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(subnameTemp, f); err != nil {
			return err
		}
	} else {
		fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(path, ".txt") {
				f, err := f.Open(path)
				if err != nil {
					return err
				}
				defer f.Close()
				if _, err := io.Copy(subnameTemp, f); err != nil {
					return err
				}
			}
			return err
		})
	}

	s.dictTempName = subnameTemp.Name()

	return err
}

func (s *SubDomain) DictList() error {
	f, err := os.Open(s.dictTempName)
	if err != nil {
		return err
	}
	defer f.Close()

	defer close(s.DictChan)

	wg := sizedwaitgroup.New(s.rateLimit)
	s2 := bufio.NewScanner(f)
	for s2.Scan() {
		wg.Add()
		go func(name string) {
			defer wg.Done()
			s.DictChan <- strings.TrimSpace(name)
		}(s2.Text())
	}
	wg.Wait()

	return nil
}
