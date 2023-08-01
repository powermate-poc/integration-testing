# PowerMate Integration testing

Integration tests run on a daily basis automatically, and validate that our cloud infrastructure is functioning as expected.

## Getting started

Tests are written in Go, using the Ginkgo testing framework.

Create an `.env` file with the following contents:

```env
HOST="https://abcdefghij.execute-api.eu-central-1.amazonaws.com"
TOKEN="Some secret token!"
BROKER="123456789-ats.iot.eu-central-1.amazonaws.com"
```