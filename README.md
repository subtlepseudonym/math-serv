## Math Server

```bash
git clone github.com/subtlepseudonym/math-serv
dep ensure
go run main.go
```

This is a simple API server that supports binary math operations.  The operations are specified via the URL path (add, subtract, multiply, etc) and variables ('x' and 'y') can be specified using a few different content types.

+ Supported math operations
	- add
	- subtract
	- multiply
	- divide
	- mod
	- pow (x to the y power)
	- root (x to the (1/y) power)
	- log (log x base y)

+ Supported content types
	- application/json
	- application/x-www-form-urlencoded

The majority of this project's content is located in the server package.  The intention there is that server can be imported seperately from the main function should someone have need of a simple binary math operations server.  