# TZO

## Dependencies
You need `Go` installed. This codebase was tested on `go1.13.5`, but doesn't use anything version specific that I'm aware of. 
```
go get github.com/dop251/goja
```

## Running TZO Server
Running a development instance of TZO is simple.
```
cd ./tzo
go run tzo-server/main.go
```

## config.json
You can specify a config file location with the `--conf` flag. By default, `./config.json` will be tried. There is an example supplied with the source.
```
--conf /path/to/config.json
```

### vm-timeout-ms
Maximum time a script may take to run.

### max-wait-ms
Maximum time that things can be waited on. For example the `.wait(ms_int)` that gets returned by asynchronous builtins like the HTTP client.

### hosts
Map of hostnames to their home directories. Scripts will be jailed inside of that directory. A special entry of `default` exists and is required.
The hosts table will be overwritten if `hosts-from-file` is specified.

### hosts-from-file
Loads hosts table from file. See `hosts.json` example.

### bind
Which network interface and port to listen over. Examples of listening on port 8080:
```
:8080
0.0.0.0:8080
127.0.0.1:8080
```

### serve-files
Default is `false`. Set to `true` if you want to serve static files. This is only recommended for development.

### script-extensions
An array of file extensions that should be treated as TZO scripts.
```
[ ".tzo", ".jx" ]
```

### X-Abuse-Info
Header to be added to outgoing HTTP requests from TZO scripts.

### net-allow-self
Default is `false`. If set to true, any requests to the server that contains the defined `X-Abuse-Info` header will be denied.

### api-password
Unused at the moment. I'm contemplating the idea of making an admin API for TZO and the password being stored here.
There's a tool for generating the `api-password` hashmap:
```
go run tools/password_gen/password_gen.go
```
The example password in the config is set to `password`.

### net-blacklist
Array of network addresses to deny network access to, in CIDR notation. Both IPv4 and IPv6 should work.


## Builtins
This should be mostly compatible with ES (ECMAScript) 5.1, since we're using `goja` as the Javascript engine. However newer ES versions will be support after migrating from `goja` to V8.

Below are the additional functions that aren't included in ES.

### print(o ...objects)
Print can take any number of arguments and print output.

### require(file string)
Executes code from specified file in the current environment.
The file path should be starting from the directory that the host is jailed to, the running script is unaware of any present working directory type of stuff.

### http_status_code(int)
Default is `200` OK. Sets http status code.

### header(name string, value string)
Sets a named header to specified string value.

### die()
Ends script execution.

### HTTP(method string, url string)
Returns a secured `HTTP_CLIENT` object.

#### HTTP_CLIENT.Request()
Returns the `HTTP_CLIENT` object itself. It initializes an HTTP request in the background to be read later.

#### HTTP_CLIENT.Wait(max_wait_ms int)
`max_wait_ms` is the maximum wait time.
Wait will pause script execution until either the time runs out or HTTP_CLIENT has received it's response.
Returns `HTTP_RESPONSE` data for script use.

#### HTTP_RESPONSE
```
type HTTP_RESPONSE struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.0"
	ProtoMajor int    // e.g. 1
	ProtoMinor int    // e.g. 0

	Header  http.Header
	Trailer http.Header

	Body          string
	ContentLength int64

    TimedOut    bool
}
```