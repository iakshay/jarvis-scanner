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
  go get -u github.com/google/gopacket

  # testing
  go get -u github.com/stretchr/testify

  # generate strings
	# optional: used to generate strings.go
  go get -u golang.org/x/tools/cmd/stringer

  # build and install worker and server
  go install ./...

  # Running worker ($GOPATH/bin/worker)
  worker

  # Running server ($GOPATH/bin/server) 
  server
```

To start the frontend server

```
  Install Node and NPM
  cd ui/
  # install dependencies
  npm install
  # start server
  npm start
```

Running Command-line Client

```
  # Building client
	cd cmd/client
	go build client.go
 
  # Sample usage of Command-line Client
  # For list jobs
	./client -task=list

  # For view specific job
	./client -task=view -id=1

  # For delete job
	./client -task=delete -id=1

  # For submit IsAlive Job
	./client -task=submit -type=IsAlive -ip=192.168.2.1/24

  # For submit PortScan Job
	./client -task=submit -type=PortScan -ip=69.63.176.0 -mode=Normal -start=0 -end=443

```
 
## Examples

IsAlive w/ IpBlock

```json
{
	"Type": 0,
	"Data": {
		"IpBlock": "192.168.2.1/24"
	}
}
```
IsAlive w/ Ip

```json
{
	"Type": 0,
	"Data": {
		"IpBlock": "192.168.2.1"
	}
}
```

PortScan

```json
{
	"Type": 1,
	"Data": {
		"Type": 1,
		"Ip": "127.0.0.1",
		 "PortRange": {
		 	"Start": 0,
		 	"End": 65535
		 }
	}
}
```
