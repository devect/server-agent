# server-agent

server-agent is a package in go to monitor your server and get information like CPU usage, Memory, Disk i/o, Network i/o... and send it to an API. It requires Golang >= 1.10


Configure the GOPATH
```
export GOPATH="/your/path/go"
```

Install dependencies:
```
go get github.com/mackerelio/go-osstat/cpu
go get github.com/mackerelio/go-osstat/memory
go get github.com/mackerelio/go-osstat/uptime
```

Run:
```
go run main.go
```

Build:
```
go1.10 build -o devect
mv devect package/
```




## Build the program

Put the binary in devect/usr/local/bin
```
mv devect package/usr/local/bin/devect
```

And execute:
```
dpkg-deb --build package/ devect.deb
```

Install the .deb package:
```
sudo dpkg -i devect.deb
```
