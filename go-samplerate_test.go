package samplerate

import "testing"
import "reflect"

func TestInitAndDestroy(t *testing.T) {
	src, err := New(SRC_SINC_FASTEST, 2)
	if err != nil {
		t.Fatal(err)
	}

	err = Delete(src)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProcessInvalidInput(t *testing.T) {
	src, err := New(SRC_SINC_FASTEST, 2)
	if err != nil {
		t.Fatal(err)
	}

	_, err = src.Process("hello", 1.0, true)
	if err == nil {
		t.Fatal("must only accept slices")
	}

	_, err = src.Process([]byte{1, 2}, 1.0, true)
	if err == nil {
		t.Fatal("expected error on invalid input type []byte")
	}

	_, err = src.Process([]int8{1, 2}, 1.0, true)
	if err == nil {
		t.Fatal("expected error on invalid input type []int8")
	}

	err = Delete(src)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProcess(t *testing.T) {
	src, err := New(SRC_SINC_FASTEST, 2)
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
