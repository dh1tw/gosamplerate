// +build linux,cgo darwin,cgo

package samplerate

/*
#cgo CFLAGS: -I /usr/local/include
#cgo LDFLAGS: -L /usr/local/lib -lsamplerate

#import <samplerate.h>
#import <stdlib.h>

SRC_DATA *alloc_src_data() {
    return (SRC_DATA*)malloc(sizeof(SRC_DATA));
}

void free_src_data(SRC_DATA *p){
    free(p);
}

*/
import "C"

import (
	"errors"
	"fmt"
	"reflect"
)

type Src struct {
	srcState *C.SRC_STATE
	channels C.long
}

const (
	SRC_SINC_BEST_QUALITY   = 0
	SRC_SINC_MEDIUM_QUALITY = 1
	SRC_SINC_FASTEST        = 2
	SRC_ZERO_ORDER_HOLD     = 3
	SRC_LINEAR              = 4
)

func New(converterType int, channels int) (Src, error) {
	cConverter := C.int(converterType)
	cChannels := C.int(channels)
	var cErr *C.int

	src_state := C.src_new(cConverter, cChannels, cErr)
	if src_state == nil {
		return Src{}, errors.New("Could not initialize")
	}

	src := Src{
		srcState: src_state,
		channels: C.long(channels),
	}

	return src, nil
}

func Delete(src Src) error {
	src_state := C.src_delete(src.srcState)
	if src_state == nil {
		return nil
	} else {
		return errors.New("Could not delete object")
	}
}

func (src *Src) Process(data interface{}, ratio float64, endOfInput bool) ([]float32, error) {

	if reflect.TypeOf(data).Kind() != reflect.Slice {
		return nil, errors.New("not a slice")
	}

	d := reflect.TypeOf(data).Elem()

	var inputLength int
	cInput := make([]C.float, 65536)
	cOutput := make([]C.float, 65536)

	switch d.Kind() {
	case reflect.Float32:
		inputLength = len(data.([]float32))

		for i, el := range data.([]float32) {
			cInput[i] = C.float(el)
		}
	case reflect.Int32:
		fmt.Println("convert to Float32")
	case reflect.Int16:
		fmt.Println("convert to Float32")
	default:
		return nil, errors.New("something else")
	}

	if inputLength == 0 {
		return nil, errors.New("empty input slice")
	}

	var cEndOfInput = C.int(0)
	if endOfInput {
		cEndOfInput = 1
	} else {
		cEndOfInput = 0
	}

	cSrcData := C.alloc_src_data()
	defer C.free_src_data(cSrcData)
	cSrcData.data_in = &cInput[0]
	cSrcData.data_out = &cOutput[0]
	cSrcData.input_frames = C.long(inputLength) / src.channels
	cSrcData.output_frames = C.long(len(cOutput)) / src.channels
	cSrcData.end_of_input = cEndOfInput
	cSrcData.src_ratio = C.double(ratio)

	res := C.src_process(src.srcState, cSrcData)

	if res != 0 {
		fmt.Println("Error code: ", res)
		return nil, errors.New("process not possible")
	}

	output := make([]float32, 0, cSrcData.output_frames_gen)

	// fmt.Println("channels:", src.channels)
	// fmt.Println("input frames", cSrcData.input_frames)
	// fmt.Println("input used", cSrcData.input_frames_used)
	// fmt.Println("output generated", cSrcData.output_frames_gen)

	for i := 0; i < int(cSrcData.output_frames_gen*src.channels); i++ {
		output = append(output, float32(cOutput[i]))
	}

	return output, nil
}
