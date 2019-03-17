# Jarvis Scanner

## Setup

```
  # Add your public key to Github
  # Setup $GOPATH
  # Clone the repository
  mkdir -p $GOPATH/src/github.com/iakshay/jarvis-scanner
  git clone git@github.com:iakshay/jarvis-scanner.git $GOPATH/src/github.com/iakshay/jarvis-scanner

  # Install dependencies
  go get -u github.com/jinzhu/gorm
  go get -u github.com/mattn/go-sqlite3

  # Running worker ($GOPATH/bin/worker)
  worker

  # Running server ($GOPATH/bin/server) 
```
