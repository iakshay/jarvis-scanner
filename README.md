## Jarvis Scanner

### Setup

```bash
  # Add your public key to Github
  # Setup $GOPATH
  # Clone the repository
  mkdir -p $GOPATH/src/github.com/iakshay/jarvis-scanner
  git clone git@github.com:iakshay/jarvis-scanner.git $GOPATH/src/github.com/iakshay/jarvis-scanner

  # Install dependencies
  
  # golang orm
  go get -u github.com/jinzhu/gorm

  # sqlite3 driver
  go get -u github.com/mattn/go-sqlite3

  # network packet related
  go get -u golang.org/x/net/ipv6
  go get -u golang.org/x/net/ipv4
  go get -u golang.org/x/net/icmp

  # testing
  go get -u github.com/stretchr/testify

  # build and install worker and server
  go install ./...

  # Running worker ($GOPATH/bin/worker)
  worker

  # Running server ($GOPATH/bin/server) 
  server
```
