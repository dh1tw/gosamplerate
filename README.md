# libsamplerate binding for Golang
[![Go Report Card](https://goreportcard.com/badge/github.com/dh1tw/gosamplerate)](https://goreportcard.com/report/github.com/dh1tw/gosamplerate)
[![Build Status](https://travis-ci.com/dh1tw/gosamplerate.svg?branch=master)](https://travis-ci.com/dh1tw/gosamplerate)
[![Coverage Status](https://coveralls.io/repos/github/dh1tw/gosamplerate/badge.svg?branch=master)](https://coveralls.io/github/dh1tw/gosamplerate?branch=master)

This is a [Golang](https://golang.org) binding for [libsamplerate](http://www.mega-nerd.com/SRC/index.html) (written in C), probably the best audio Sample Rate Converter available to today.

A classical use case is converting audio from a CD sample rate of 44.1kHz to the 48kHz sample rate used by DAT players.

libsamplerate is capable of arbitrary and time varying conversions (max sampling / upsampling by factor 256) and comes with 5 converters, allowing quality to be traded off against computation cost.

## API implementations
**gosamplerate** implements the following [libsamplerate](http://www.mega-nerd.com/SRC/index.html) API calls:

- [Simple API](http://www.mega-nerd.com/SRC/api_simple.html)
- [Full API](http://www.mega-nerd.com/SRC/api_full.html)
- [Most miscellaneous functions](http://www.mega-nerd.com/SRC/api_misc.html)

not (yet) implemented is:

- [Callback API](http://www.mega-nerd.com/SRC/api_callback.html)

## License
This library (**gosamplerate**) is published under the the permissive [BSD license](http://choosealicense.com/licenses/mit/). You can find a good comparison of Open Source Software licenses, including the BSD license at [choosealicense.com](http://choosealicense.com/licenses/)

libsamplerate has been [republished in 2016 under the 2-clause BSD license](http://www.mega-nerd.com/SRC/license.html).

## How to install samplerate

Make sure that you have [libsamplerate](http://www.mega-nerd.com/SRC/index.html) installed on your system.

On Mac or Linux it can be installed conveniently through your distro's packet manager.

### Linux:
using **apt** (Ubuntu), **yum** (Centos)...etc.
```bash
    $ sudo apt install libsamplerate0
```

### MacOS
using [Homebrew](http://brew.sh):
```bash
    $ brew install libsamplerate
```

### Install gosamplerate
```bash
    $ go get github.com/dh1tw/gosamplerate
```

## Documentation
The API of **gosamplerate** can be found at [godoc.org](https://godoc.org/github.com/dh1tw/gosamplerate).
The documentation of libsamplerate (necessary in order to fully understand the API) can be found
[here](http://www.mega-nerd.com/SRC/index.html).

## Tests & Examples
The test coverage is close to 100%. The tests contain various examples on how to use **gosamplerate**.