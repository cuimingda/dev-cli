package cmd

import (
	"fmt"
	"io"
	"time"
)

type TimezoneRunner struct {
	now func() time.Time
}

func newTimezoneRunner() *TimezoneRunner {
	return &TimezoneRunner{
		now: time.Now,
	}
}

func (r *TimezoneRunner) Run(output io.Writer) error {
	_, err := fmt.Fprintln(output, formatLocalTimezoneDisplay(r.now()))
	return err
}
