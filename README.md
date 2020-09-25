[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/monch1962/async-pact-spike)
[![Build Status](https://dev.azure.com/monch1962/monch1962/_apis/build/status/monch1962.async-pact-spike?branchName=master)](https://dev.azure.com/monch1962/monch1962/_build/latest?definitionId=13&branchName=master)

# async-pact-spike
Spike to play with Pact testing on RabbitMQ

## Environment variables

There are several environment variables that can be configured for this tool, some of which are optional:
- `PROTOCOL` (required, can only be `amqp` or `amqps`)
- `PUBLISH_AMQP_SERVER` (default: `localhost`)
- `PUBLISH_AMQP_SERVER_TCP` (default: `5672` if PROTOCOL is `amqp`, or `5671` if PROTOCOL is `amqps`)
- `PUBLISH_USERNAME` (default: `guest`)
- `PUBLISH_PASSWORD` (default: `guest`)
- `PUBLISH_Q` (required)
- `SUBSCRIBE_AMQP_SERVER` (default: same as `PUBLISH_AMQP_SERVER`)
- `SUBSCRIBE_AMQP_SERVER_TCP` (default: same as `SUBSCRIBE_AMQP_SERVER_TCP`)
- `SUBSCRIBE_USERNAME` (default: same as `PUBLISH_USERNAME`)
- `SUBSCRIBE_PASSWORD` (default: same as `PUBLISH_PASSWORD`)
- `SUBSCRIBE_Q` (required)

## To run
To run tests against a AMQP server on localhost in default config:

`$ PUBLISH_Q=abc SUBSCRIBE_Q=def go test`

or to specify AMQP server details, use something like:

`$ PUBLISH_AMQP_SERVER=peacock.rmq.cloudamqp.com PUBLISH_USERNAME=edtdqjib PUBLISH_PASSWORD=ucsUTMSeTR9lQEo0cRJwCgPicoroEPwa PUBLISH_Q=abc123 PUBLISH_URI_SUFFIX=edtdqjib PUBLISH_AMQP_SERVER_TCP=5671 PROTOCOL=amqps SUBSCRIBE_Q=abc123 TIMEOUT=10000 go test`

## To output JUnit format

`$ go get -u github.com/jstemmer/go-junit-report`

`$ PUBLISH_AMQP_SERVER=peacock.rmq.cloudamqp.com PUBLISH_USERNAME=edtdqjib PUBLISH_PASSWORD=ucsUTMSeTR9lQEo0cRJwCgPicoroEPwa PUBLISH_Q=abc123 PUBLISH_URI_SUFFIX=edtdqjib PUBLISH_AMQP_SERVER_TCP=5671 PROTOCOL=amqps SUBSCRIBE_Q=abc123 TIMEOUT=10000 go test -v 2>&1 | go-junit-report`

## To compile tests into a standalone executable file

`$ go test -c -o tests`

will compile all the tests into a standalone EXE called `tests`, which can be moved and executed on other hardware

`$ PUBLISH_AMQP_SERVER=peacock.rmq.cloudamqp.com PUBLISH_USERNAME=edtdqjib PUBLISH_PASSWORD=ucsUTMSeTR9lQEo0cRJwCgPicoroEPwa PUBLISH_Q=abc123 PUBLISH_URI_SUFFIX=edtdqjib PUBLISH_AMQP_SERVER_TCP=5671 PROTOCOL=amqps SUBSCRIBE_Q=abc123 TIMEOUT=10000 ./tests -test.v` 

will run the tests in verbose mode, and 

`$ PUBLISH_AMQP_SERVER=peacock.rmq.cloudamqp.com PUBLISH_USERNAME=edtdqjib PUBLISH_PASSWORD=ucsUTMSeTR9lQEo0cRJwCgPicoroEPwa PUBLISH_Q=abc123 PUBLISH_URI_SUFFIX=edtdqjib PUBLISH_AMQP_SERVER_TCP=5671 PROTOCOL=amqps SUBSCRIBE_Q=abc123 TIMEOUT=10000 ./tests` 

will run them without verbose mode.

Finally, to generate Junit reports from a compiled test file, 

`$ PUBLISH_AMQP_SERVER=peacock.rmq.cloudamqp.com PUBLISH_USERNAME=edtdqjib PUBLISH_PASSWORD=ucsUTMSeTR9lQEo0cRJwCgPicoroEPwa PUBLISH_Q=abc123 PUBLISH_URI_SUFFIX=edtdqjib PUBLISH_AMQP_SERVER_TCP=5671 PROTOCOL=amqps SUBSCRIBE_Q=abc123 TIMEOUT=10000 ./tests -test.v | go-junit-report`