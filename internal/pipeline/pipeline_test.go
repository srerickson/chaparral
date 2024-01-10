package pipeline_test

import (
	"errors"
	"testing"

	"github.com/srerickson/chaparral/internal/pipeline"
)

func TestPipelineNil(t *testing.T) {
	err := pipeline.Run[job, result](nil, nil, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPipelineSetupErr(t *testing.T) {
	input := func(add func(job) bool) error {
		return errors.New("catch me")
	}
	output := func(in job, out result, err error) error {
		return err
	}
	err := pipeline.Run(input, doWork, output, 0)
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestPipelineResultErr(t *testing.T) {
	input := func(add func(job) bool) error {
		add(job(-1)) // invalid job
		return nil
	}
	output := func(in job, out result, err error) error {
		return err
	}
	err := pipeline.Run(input, doWork, output, 0)
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestPipeline(t *testing.T) {
	times := 100
	input := func(add func(job) bool) error {
		for i := 0; i < times; i++ {
			add(job(i))
		}
		return nil
	}
	var results int
	output := func(in job, out result, err error) error {
		results++
		return nil
	}
	err := pipeline.Run(input, doWork, output, 0)
	if err != nil {
		t.Fatal(err)
	}
	if results != times {
		t.Fatalf("output func ran %d times, not %d", results, times)
	}
}

func TestPipelineCancel(t *testing.T) {
	setupErr := errors.New("setup terminated")
	resultErr := errors.New("result error")
	input := func(add func(job) bool) error {
		i := 0
		for {
			i++
			if !add(job(i)) {
				return setupErr
			}
		}
	}
	times := 0
	output := func(in job, out result, err error) error {
		times++
		if times > 100 {
			return resultErr
		}
		return nil
	}
	err := pipeline.Run(input, doWork, output, 0)
	if err == nil {
		t.Fatal("Run() didn't return error returne from setup and result")
	}
	if !errors.Is(err, setupErr) {
		t.Error("Run() error didn't include error from setup")
	}
	if !errors.Is(err, resultErr) {
		t.Error("Run() error didn't include error from result")
	}
}

func BenchmarkPipeline(b *testing.B) {
	input := func(add func(job) bool) error {
		for i := 0; i < b.N; i++ {
			if !add(job(i)) {
				return errors.New("job not added")
			}
		}
		return nil
	}
	output := func(in job, out result, err error) error {
		return nil
	}
	b.ResetTimer()
	b.ReportAllocs()
	err := pipeline.Run(input, doWork, output, 0)
	if err != nil {
		b.Fatal(err)
	}
}

type job int
type result int

func doWork(j job) (result, error) {
	var r result
	if j < 0 {
		return r, errors.New("invalid value")
	}
	r = result(j * 2)
	return r, nil
}
