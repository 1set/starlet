# :sparkles: Starlet - Supercharging Starlark, Simply

[![godoc](https://pkg.go.dev/badge/github.com/1set/starlet.svg)](https://pkg.go.dev/github.com/1set/starlet)
[![codecov](https://codecov.io/github/1set/starlet/branch/master/graph/badge.svg?token=M1tauam4Hw)](https://codecov.io/github/1set/starlet)
[![codacy](https://app.codacy.com/project/badge/Grade/4e9c3f67a9574e6caa1b0d4706535815)](https://app.codacy.com/gh/1set/starlet/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![codeclimate](https://api.codeclimate.com/v1/badges/290ec0cc3261d16c423f/maintainability)](https://codeclimate.com/github/1set/starlet/maintainability)
[![goreportcard](https://goreportcard.com/badge/github.com/1set/starlet)](https://goreportcard.com/report/github.com/1set/starlet)

> Enhancing your Starlark scripting experience with powerful extensions and enriched wrappers. Start your Starlark journey with Starlet, where simplicity meets functionality.

Starlet is yet another Go wrapper for the official [*Starlark in Go*](https://github.com/google/starlark-go) project, designed to enhance the Starlark scripting experience with powerful extensions and enriched wrappers, and provide a more user-friendly and powerful interface for embedding Starlark scripting in your Go applications.

Inspired by the [*Starlight*](https://github.com/starlight-go/starlight) and [*Starlib*](https://github.com/qri-io/starlib) projects, Starlet focuses on three main objectives: providing an easier-to-use **Go wrapper** for Starlark, offering seamless **data conversion** between Go and Starlark, and supplying a set of **useful libraries** for Starlark.

## Key Features

*Starlet* provides the following key features:

### Flexible Machine Abstraction

*Starlet* introduces a streamlined interface for executing Starlark scripts, encapsulating the complexities of setting up and running the scripts --- The [`Machine`](https://pkg.go.dev/github.com/1set/starlet#Machine) type in *Starlet* serves as a comprehensive wrapper for Starlark runtime environments, offering an intuitive API for executing Starlark scripts, managing global variables, loading modules, controlling the script execution flow., and handling script outputs.

### Enhanced Data Conversion

*Starlet* offers the [`dataconv`](https://pkg.go.dev/github.com/1set/starlet/dataconv) package that simplifies the data exchange between Go and Starlark types. Unlike *Starlight*, which wraps Go values in Starlark-friendly structures, it focuses on transforming Go values into their Starlark equivalents and vice versa. This allows for a more seamless integration of Go's rich data types into Starlark scripts.

### Extended Libraries & Functionalities

*Starlet* includes a set of custom modules and libraries that extend the functionality of the Starlark language. These modules cover a wide range of use cases, such as file manipulation, HTTP client, JSON/CSV handling, and more, making Starlark scripts even more powerful and versatile.

## Libraries

Starlet offers a variety of libraries that extend the functionality of Starlark. Below is a brief introduction of each library:

| Package                           | Go Docs                                                                                                                                      | Description                                                   |
|:----------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------|:--------------------------------------------------------------|
| [`atom`](/lib/atom)               | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/atom.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/atom)               | Atomic operations for integers, floats, and strings           |
| [`base64`](/lib/base64)           | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/base64.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/base64)           | Base64 encoding & decoding functions                          |
| [`csv`](/lib/csv)                 | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/csv.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/csv)                 | Parses and writes comma-separated values (csv) contents       |
| [`file`](/lib/file)               | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/file.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/file)               | Functions to interact with the file system                    |
| [`goidiomatic`](/lib/goidiomatic) | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/goidiomatic.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/goidiomatic) | Go idiomatic functions and values for Starlark                |
| [`hashlib`](/lib/hashlib)         | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/hashlib.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/hashlib)         | Hash primitives for Starlark                                  |
| [`http`](/lib/http)               | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/http.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/http)               | HTTP client and server handler implementation for Starlark    |
| [`json`](/lib/json)               | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/json.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/json)               | Utilities for converting Starlark values to/from JSON strings |
| [`log`](/lib/log)                 | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/log.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/log)                 | Functionality for logging messages at various severity levels |
| [`path`](/lib/path)               | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/path.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/path)               | Functions to manipulate directories and file paths            |
| [`random`](/lib/random)           | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/random.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/random)           | Functions to generate random values for various distributions |
| [`re`](/lib/re)                   | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/re.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/re)                   | Regular expression functions for Starlark                     |
| [`runtime`](/lib/runtime)         | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/runtime.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/runtime)         | Provides Go and app runtime information                       |
| [`string`](/lib/string)           | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/string.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/string)           | Constants and functions to manipulate strings                 |
