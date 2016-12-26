// +build linux,cgo darwin,cgo

package gosamplerate

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
)

type Src struct {
	srcState     *C.SRC_STATE
	srcData      *C.SRC_DATA
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

//New initializes the converter object and returns a reference to it.
func New(converterType int, channels int, buffer_len int) (Src, error) {
	cConverter := C.int(converterType)
	cChannels := C.int(channels)
	var cErr *C.int

	src_state := C.src_new(cConverter, cChannels, cErr)
	if src_state == nil {
		return Src{}, errors.New("Could not initialize samplerate converter object")
	}

	inputBuffer := make([]C.float, buffer_len)
	outputBuffer := make([]C.float, buffer_len)

	cData := C.alloc_src_data()
	cData.data_in = &inputBuffer[0]
	cData.data_out = &outputBuffer[0]
	cData.output_frames = C.long(len(outputBuffer) / channels)
	cData.src_ratio = 1

	src := Src{
		srcState:     src_state,
		srcData:      cData,
		channels:     C.long(cChannels),
		inputBuffer:  inputBuffer,
		outputBuffer: outputBuffer,
	}

	return src, nil
}

// Delete cleans up all internal allocations.
func Delete(src Src) error {
	srcState := C.src_delete(src.srcState)
	C.free_src_data(src.srcData)
	if srcState == nil {
		return nil
	}
	return errors.New("Could not delete object; It did not exist")
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
		errMsg := fmt.Sprintf("Could not reset samplerate converter state: %s",
			Error(int(res)))
		return errors.New(errMsg)
	}
	return nil
}

// Error Convert the error number into a string.
func Error(errNo int) string {
	err := C.src_strerror(C.int(errNo))
	return C.GoString(err)
}

//ErrorNo return an error number
func (src *Src) ErrorNo() int {
	errNo := C.src_error(src.srcState)
	return int(errNo)
}

// Simple converts a single block of samples (one or more channels) in one go. The simple API is less
// capable than the full API (Process()). It must not be used if Audio shall be converted in junks. For full documentation see: http://www.mega-nerd.com/SRC/api_simple.html
func Simple(dataIn []float32, ratio float64, channels int, converterType int) ([]float32, error) {
	cConverterType := C.int(converterType)
	cChannels := C.int(channels)
	cRatio := C.double(ratio)

	dataInLength := len(dataIn)

	inputBuffer := make([]C.float, dataInLength)
	outputBuffer := make([]C.float, dataInLength*int(ratio)+20) // add some margin

	// copy data into input buffer
	for i, el := range dataIn {
		inputBuffer[i] = C.float(el)
	}

	srcData := C.alloc_src_data()
	defer C.free_src_data(srcData)

	srcData.data_in = &inputBuffer[0]
	srcData.data_out = &outputBuffer[0]
	srcData.input_frames = C.long(dataInLength / channels)
	srcData.output_frames = C.long(cap(outputBuffer) / channels)
	srcData.src_ratio = cRatio

	res := C.src_simple(srcData, cConverterType, cChannels)

	if res != 0 {
		errMsg := fmt.Sprintf("Error code: %d; %s", res, Error(int(res)))
		return nil, errors.New(errMsg)
	}

	output := make([]float32, 0, srcData.output_frames_gen)

	// fmt.Println("channels:", src.channels)
	// fmt.Println("input frames", cSrcData.input_frames)
	// fmt.Println("input used", cSrcData.input_frames_used)
	// fmt.Println("output generated", cSrcData.output_frames_gen)

	for i := 0; i < int(srcData.output_frames_gen*C.long(channels)); i++ {
		output = append(output, float32(outputBuffer[i]))
	}

	return output, nil
}

// Process is known as the full API. It allows time varying sample rate conversion on streaming data on one or more channels. For full documentation see: http://www.mega-nerd.com/SRC/api_full.html
func (src *Src) Process(dataIn []float32, ratio float64, endOfInput bool) ([]float32, error) {

	inputLength := len(dataIn)

	if inputLength > len(src.inputBuffer) {
		return nil, errors.New("data slice is larger than buffer")
	}

	// copy data into input buffer
	for i, el := range dataIn {
		src.inputBuffer[i] = C.float(el)
	}

	var cEndOfInput = C.int(0)
	if endOfInput {
		cEndOfInput = 1
	} else {
		cEndOfInput = 0
	}

	src.srcData.input_frames = C.long(inputLength) / src.channels
	src.srcData.end_of_input = cEndOfInput
	src.srcData.src_ratio = C.double(ratio)

	res := C.src_process(src.srcState, src.srcData)

	if res != 0 {
		errMsg := fmt.Sprintf("Error code: %d; %s", res, Error(int(res)))
		return nil, errors.New(errMsg)
	}

	output := make([]float32, 0, src.srcData.output_frames_gen)

	// fmt.Println("channels:", src.channels)
	// fmt.Println("input frames", cSrcData.input_frames)
	// fmt.Println("input used", cSrcData.input_frames_used)
	// fmt.Println("output generated", cSrcData.output_frames_gen)

	for i := 0; i < int(src.srcData.output_frames_gen*src.channels); i++ {
		output = append(output, float32(src.outputBuffer[i]))
	}

	return output, nil
}

// SetRatio sets the samplerate conversion ratio between input and output samples.
// Normally, when using (src *SRC) Process or the callback, the libary will try to smoothly
// transition between the conversion ratio of the last call and the conversion ratio of the next
// call. This function bypasses this behaviour and achieves a step response in the conversion rate.
func (src *Src) SetRatio(ratio float64) error {

	res := C.src_set_ratio(src.srcState, C.double(ratio))
	if res != 0 {
		return errors.New(Error(int(res)))
	}
	return nil
}

// IsValidRatio returns True is ratio is a valid conversion ratio, False otherwise.
func IsValidRatio(ratio float64) bool {
	res := C.src_is_valid_ratio(C.double(ratio))
	if res == 1 {
		return true
	}
	return false
}

// GetName returns the name of a sample rate converter
func GetName(converter C.int) (string, error) {
	cConverterName := C.src_get_name(converter)
	if cConverterName == nil {
		return "", errors.New("unknown samplerate converter")
	}
	return C.GoString(cConverterName), nil
}

// GetDescription returns the description of a sample rate converter
func GetDescription(converter C.int) (string, error) {
	cConverterDescription := C.src_get_description(converter)
	if cConverterDescription == nil {
		return "", errors.New("unknown samplerate converter")
	}
	return C.GoString(cConverterDescription), nil
}

// GetVersion returns the version number of libsamplerate
func GetVersion() string {

	cVersion := C.src_get_version()
	return C.GoString(cVersion)
}
