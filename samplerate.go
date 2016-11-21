// +build linux,cgo darwin,cgo

package samplerate

/*
#cgo CFLAGS: -I /usr/local/include
#cgo LDFLAGS: -L /usr/local/lib -lsamplerate

#include <samplerate.h>
#include <stdlib.h>

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
	srcState     *C.SRC_STATE
	channels     C.long
	inputBuffer  []C.float
	outputBuffer []C.float
}

const (
	SRC_SINC_BEST_QUALITY   = C.SRC_SINC_BEST_QUALITY
	SRC_SINC_MEDIUM_QUALITY = C.SRC_SINC_MEDIUM_QUALITY
	SRC_SINC_FASTEST        = C.SRC_SINC_FASTEST
	SRC_ZERO_ORDER_HOLD     = C.SRC_ZERO_ORDER_HOLD
	SRC_LINEAR              = C.SRC_LINEAR
)

func New(converterType int, channels int, buffer_len int) (Src, error) {
	cConverter := C.int(converterType)
	cChannels := C.int(channels)
	var cErr *C.int

	src_state := C.src_new(cConverter, cChannels, cErr)
	if src_state == nil {
		return Src{}, errors.New("Could not initialize")
	}

	src := Src{
		srcState:     src_state,
		channels:     C.long(cChannels),
		inputBuffer:  make([]C.float, buffer_len),
		outputBuffer: make([]C.float, buffer_len),
	}

	return src, nil
}

// Delete cleans up all internal allocations.
func Delete(src Src) error {
	srcState := C.src_delete(src.srcState)
	if srcState == nil {
		return nil
	}
	return errors.New("Could not delete object; It does not exist")
}

// GetChannels gets the current channel count.
func (src *Src) GetChannels() (int, error) {
	// for version < 1.9
	return int(src.channels), nil

	// with version 1.9 src_get_channels was added
	// cChannels := C.src_get_channels(src.srcState)
	// if cChannels < 0 {
	// 	return int(cChannels), errors.New("invalid channel count")
	// }
	// return int(cChannels), nil
}

// Reset the internal SRC state. It does not modify the quality settings.
// It does not free any memory allocations.
func (src *Src) Reset() error {
	res := C.src_reset(src.srcState)
	if res < 0 {
		return errors.New("could not reset state")
	}
	return nil
}

// Error Convert the error number into a string.
func (src *Src) Error(errNo int) string {
	err := C.src_strerror(C.int(errNo))
	return C.GoString(err)
}

//ErrorNo return an error number
func (src *Src) ErrorNo() int {
	errNo := C.src_error(src.srcState)
	return int(errNo)
}

func (src *Src) Process(data interface{}, ratio float64, endOfInput bool) ([]float32, error) {

	if reflect.TypeOf(data).Kind() != reflect.Slice {
		return nil, errors.New("not a slice")
	}

	d := reflect.TypeOf(data).Elem()

	var inputLength int

	switch d.Kind() {
	case reflect.Float32:
		inputLength = len(data.([]float32))

		for i, el := range data.([]float32) {
			src.inputBuffer[i] = C.float(el)
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
	cSrcData.data_in = &src.inputBuffer[0]
	cSrcData.data_out = &src.outputBuffer[0]
	cSrcData.input_frames = C.long(inputLength) / src.channels
	cSrcData.output_frames = C.long(len(src.outputBuffer)) / src.channels
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
		output = append(output, float32(src.outputBuffer[i]))
	}

	return output, nil
}

// GetName returns the name of a sample rate converter
func GetName(converter C.int) (string, error) {
	cConverterName := C.src_get_name(converter)
	if cConverterName == nil {
		return "", errors.New("unknown converter")
	}
	return C.GoString(cConverterName), nil
}

// GetDescription returns the description of a sample rate converter
func GetDescription(converter C.int) (string, error) {
	cConverterDescription := C.src_get_description(converter)
	if cConverterDescription == nil {
		return "", errors.New("unknown converter")
	}
	return C.GoString(cConverterDescription), nil
}

// GetVersion returns the version number of libsamplerate
func GetVersion() string {
	cVersion := C.src_get_version()
	return C.GoString(cVersion)
}

// IsValidRatio returns True is ratio is a valid conversion ratio, False otherwise.
func IsValidRatio(ratio float64) bool {
	res := C.src_is_valid_ratio(C.double(ratio))
	if res == 1 {
		return true
	}
	return false
}
