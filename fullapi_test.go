package samplerate

import (
	"reflect"
	"testing"
)

func TestInitAndDestroy(t *testing.T) {
	channels := 2
	src, err := New(SRC_SINC_FASTEST, channels, 100)
	if err != nil {
		t.Fatal(err)
	}

	chs, err := src.GetChannels()
	if err != nil {
		t.Fatal(err)
	}
	if chs != channels {
		t.Fatal("unexpected amount of channels")
	}

	err = src.Reset()
	if err != nil {
		t.Fatal(err)
	}

	err = Delete(src)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInvalidSrcObject(t *testing.T) {
	_, err := New(5, 2, 100)
	if err == nil {
		t.Fatal("expected Error")
	}
	if err.Error() != "Could not initialize samplerate converter object" {
		t.Log("unexpected Error string")
	}
}

func TestProcess(t *testing.T) {
	src, err := New(SRC_LINEAR, 2, 100)
	if err != nil {
		t.Fatal(err)
	}

	input := []float32{0.1, -0.5, 0.2, -0.3}
	output, err := src.Process(input, 2.0, false)
	if err != nil {
		t.Fatal(err)
	}
	expOutput := []float32{0.1, -0.5, 0.1, -0.5, 0.1, -0.5, 0.15, -0.4}

	if !reflect.DeepEqual(output, expOutput) {
		t.Log("input:", input)
		t.Log("output:", output)
		t.Fatal("unexpected output")
	}

	err = Delete(src)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProcessWithEndOfInputFlagSet(t *testing.T) {
	src, err := New(SRC_SINC_FASTEST, 2, 100)
	if err != nil {
		t.Fatal(err)
	}

	input := []float32{0.1, -0.5, 0.2, -0.3}
	output, err := src.Process(input, 2.0, true)
	if err != nil {
		t.Fatal(err)
	}
	expOutput := []float32{0.11488709,
		-0.46334597, 0.18373828, -0.48996875, 0.1821644,
		-0.32879135, 0.10804618, -0.11150829}

	if !reflect.DeepEqual(output, expOutput) {
		t.Log("input:", input)
		t.Log("output:", output)
		t.Fatal("unexpected output")
	}

	err = Delete(src)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProcessDataSliceBiggerThanInputBuffer(t *testing.T) {
	src, err := New(SRC_LINEAR, 1, 100)
	if err != nil {
		t.Fatal(err)
	}

	input := make([]float32, 150)
	_, err = src.Process(input, 150.0, true)
	if err == nil {
		t.Fatal("expected Error")
	}
	if err.Error() != "data slice is larger than buffer" {
		t.Log("unexpected Error string")
	}
}

func TestProcessErrorWithInvalidRatio(t *testing.T) {
	src, err := New(SRC_LINEAR, 1, 100)
	if err != nil {
		t.Fatal(err)
	}

	input := make([]float32, 100)
	_, err = src.Process(input, -5, true)
	if err == nil {
		t.Fatal("expected Error")
	}
	if err.Error() != "Error code: 6; SRC ratio outside [1/256, 256] range." {
		t.Log(err.Error())
		t.Log("unexpected Error string")
	}
}

func TestGetChannels(t *testing.T) {
	channels := 2
	src, err := New(SRC_SINC_FASTEST, channels, 100)
	if err != nil {
		t.Fatal(err)
	}
	chLength, err := src.GetChannels()
	if err != nil {
		t.Fatal(err)
	} else if chLength != channels {
		t.Fatal("unexpected channel length")
	}
}

func TestSetRatio(t *testing.T) {
	src, err := New(SRC_LINEAR, 1, 10)
	if err = src.SetRatio(25.0); err != nil {
		t.Fatal("unexpected result; should be valid conversion rate")
	}
}

func TestSetRatioInvalid(t *testing.T) {
	src, err := New(SRC_LINEAR, 1, 10)
	err = src.SetRatio(-5)
	if err == nil {
		t.Fatal("expected Error")
	}
}

func TestIsValidRatio(t *testing.T) {
	if !IsValidRatio(5) {
		t.Fatal("unexpected result; should be valid")
	}

	if IsValidRatio(-1) {
		t.Fatal("unexpected result; should be invalid")
	}

	if IsValidRatio(257) {
		t.Fatal("unexpected result; should be invalid")
	}
}

func TestErrors(t *testing.T) {
	channels := 2
	src, err := New(SRC_SINC_FASTEST, channels, 100)
	if err != nil {
		t.Fatal(err)
	}

	errNo := src.ErrorNo()
	if errNo != 0 {
		t.Fatal("unexpected error number")
	}

	errString := src.Error(0)
	if errString != "No error." {
		t.Fatal("unexpected Error string")
	}

	err = Delete(src)
	if err != nil {
		t.Fatal(err)
	}
}
