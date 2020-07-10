// Package gosamplerate is a golang binding for libsamplerate (audio sample rate converter)
package gosamplerate

/*

#cgo pkg-config: samplerate

#include <samplerate.h>
#include <stdlib.h>

SRC_DATA *alloc_src_data(float *data_in, float *data_out,
					long output_frames, double src_ratio) {

	SRC_DATA *src_data = malloc(sizeof(SRC_DATA));
    src_data->data_in = data_in;
    src_data->data_out = data_out;
    src_data->output_frames = output_frames;
    src_data->src_ratio = src_ratio;
	return src_data;
}

void free_src_data(SRC_DATA *p){
    free(p);
}

int run_src_simple(float *data_in, float *data_out,
                   long input_frames, long output_frames,
                   double src_ratio, int converter_type, int channels,
                   long *output_frames_gen) {
    SRC_DATA src_data;
    src_data.data_in = data_in;
    src_data.data_out = data_out;
    src_data.input_frames = input_frames;
    src_data.output_frames = output_frames;
    src_data.src_ratio = src_ratio;
    int res = src_simple(&src_data, converter_type, channels);
    *output_frames_gen = src_data.output_frames_gen;
    return res;
}

*/
import "C"

import (
	"errors"
	"fmt"
	"math"
)

// Src struct holding the data for the full API
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
func New(converterType int, channels int, bufferLen int) (Src, error) {
	cConverter := C.int(converterType)
	cChannels := C.int(channels)
	var cErr *C.int

	src_state := C.src_new(cConverter, cChannels, cErr)
	if src_state == nil {
		return Src{}, errors.New("Could not initialize samplerate converter object")
	}

	inputBuffer := make([]C.float, bufferLen)
	outputBuffer := make([]C.float, bufferLen)

	cData := C.alloc_src_data(
		&inputBuffer[0],
		&outputBuffer[0],
		C.long(len(outputBuffer)/channels),
		1,
	)

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
// capable than the full API (Process()). It must not be used if Audio shall be converted in chunks. For full documentation see: http://www.mega-nerd.com/SRC/api_simple.html
func Simple(dataIn []float32, ratio float64, channels int, converterType int) ([]float32, error) {
	if ratio <= 0 {
		return nil, fmt.Errorf("Error code: 6; SRC ratio outside [1/256, 256] range.")
	}
	outputBuffer := make([]float32, len(dataIn)*int(math.Ceil(ratio)))
	var outputFramesGen C.long
	res := C.run_src_simple(
		(*C.float)(&dataIn[0]),
		(*C.float)(&outputBuffer[0]),
		C.long(len(dataIn)/channels),
		C.long(cap(outputBuffer)/channels),
		C.double(ratio),
		C.int(converterType),
		C.int(channels),
		&outputFramesGen)
	if res != 0 {
		errMsg := fmt.Sprintf("Error code: %d; %s", res, Error(int(res)))
		return nil, errors.New(errMsg)
	}
	return outputBuffer[:int(outputFramesGen)*channels], nil
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
// Normally, when using (src *SRC) Process or the callback, the library will try to smoothly
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
