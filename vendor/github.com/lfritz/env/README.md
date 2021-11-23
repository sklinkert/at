# env

[![Build Status](https://travis-ci.org/lfritz/env.svg?branch=master)](https://travis-ci.org/lfritz/env)
[![GoDoc](https://godoc.org/github.com/lfritz/env?status.svg)](https://godoc.org/github.com/lfritz/env)
[![Go Report Card](https://goreportcard.com/badge/github.com/lfritz/env)](https://goreportcard.com/report/github.com/lfritz/env)

A library for loading environment variables

Loading configuration from environment variables with `os.Getenv` can get tedious. This package
provides an interface to specify expected environment variables and takes care of loading the values
and producing good error messages if a variable is missing or has the wrong format. It can also
generate documentation and it has a way to define groups of variables with a common prefix.


## Usage

You first create an `env.Env` and call its methods to set expected environment variables:

```
var config struct{
	name    string
	port    int
	verbose bool
}
e := env.New()
e.String("NAME", &config.name, "a short name for the service")
e.OptionalInt("PORT", &config.port, 8080, "the port number to listen on (default: 8080)")
e.Bool("VERBOSE", &config.verbose, "produce verbose error messages")
```

Then, `Load` actually loads the variables:

```
err := e.Load()
if err != nil {
	fmt.Printf("error: %s", e)
}
```

The `Help` method produces documentation in this format:

```
NAME -- a short name for the service
PORT -- the port number to listen on
VERBOSE -- produce verbose error messages
```

The `FromMap` function let you hard-code an environment, which can be useful for testing.

Finally, the `Prefix` method lets you define a set of variables with a common prefix. It returns a
new `Env` instance that can be passed around, so you can move the configuration code closer to where
the configuration is needed.
