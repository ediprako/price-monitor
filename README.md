# go-price-monitor

Prerequisites
- Docker
- free port 5432 for database, 8080 for application

## How to Run this project
Since the project already use Go Module, I recommend to put the source code in any folder but GOPATH.

## Run testing
```bash
$ make test
```

## Run the application
```bash
### move to directory
$ cd workspace

### run application
$ make run

### testing application
$ curl localhost:8080/ping

### stop application
$ make stop

```

Or yo can test application via browser at [`http://localhost:8080`](http://localhost:8080)

# Tools used
In this project, i use some tools / library that listed at [`go.mod`](https://github.com/ediprako/price-monitor/blob/master/go.mod) 
