[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/monch1962/async-pact-spike)

# async-pact-spike
Spike to play with Pact testing on RabbitMQ

## Environment variables

There are several environment variablesthat can be configured for this tool, some of which are optional:
- `PUBLISH_AMQP_SERVER` (default: `localhost`)
- `PUBLISH_AMQP_SERVER_TCP` (default: `5672`)
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