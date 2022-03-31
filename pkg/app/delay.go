package app

import (
	"time"

	"github.com/ohler55/ojg/oj"
)

type Delayer interface {
	Apply(*Delay)
}

type Delay struct {
	Fixed FixedDelay `json:"fixed"`
}

type FixedDelay struct {
	Duration Duration `json:"duration"`
}

type ResponseDelayer struct{}

func NewResponseDelayer() ResponseDelayer {
	return ResponseDelayer{}
}

func (ResponseDelayer) Apply(delay *Delay) {
	if delay == nil {
		return
	}

	if delay.Fixed.Duration != 0 {
		time.Sleep(time.Duration(delay.Fixed.Duration))
	}
}

type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	err := oj.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(dur)
	return nil
}
