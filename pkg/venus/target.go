package venus

import (
	fileutil "github.com/zan8in/pins/file"
)

func (r *Runner) PreprocessTarget() error {
	var (
		err error
	)

	if len(r.options.TargetFile) > 0 {

		flist, err := fileutil.ReadFile(r.options.TargetFile)
		if err != nil {
			return err
		}

		for f := range flist {
			r.options.Target.Set(f)
		}
	}

	return err
}
