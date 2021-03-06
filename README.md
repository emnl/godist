# Godist
**– _Transmission control protocol (TCP) distribution_**

	|      | --TCP--> |      | --TCP--> |       |
	|Client|          |Godist|          |Service|
	|      | <--TCP-- |      | <--TCP-- |       |

Godist distributes your incoming tcp connections between servers, balancing the load. A thin high-performance layer between your servers and the world.

Written in Go. Built for concurrency.

## Architecture

Godist spawn **1** new Goroutine for each incomming connection, the "handler".

	Incomming
	connections
	    ->              -> New handler (TCP conn)
	    ->       Server -> New handler (TCP conn)
	    ->              -> New handler (TCP conn)

The Goroutine then connect to a server based on the clients ip, it then spawns **2** more Goroutines to pass data recieved from the client going to the server, and the response from the server going to the client.

             -> passData from:client to:server
	Handler |
	         -> passData from:server to:client

One incomming connection requires **3** Goroutines and 2 file descriptors (FD).

The buffer size for each "passData" is set at 2048 bytes.

## Usage
Godist requires a configuration file to setup the most basic settings. The software will look for **godist.conf** by default. You can specify your own configuration file by passing it as an argument.

**godist.conf**

    {
		"Host": "localhost:8080",
		"Servers":
			["localhost:4000", 
			 "localhost:4001",
			 "localhost:4002"]
	}

Then:

    $ godist godist.conf

### Distributing connections
The incoming connections are distributed by hashing the ip and port. An ip and port will always get the same server handling its connection.

## Build & Install
Godist is exceptionally easy to build and install. Just clone the repo and execute one of the following commands.

    $ go build

Will get you a ./**godist** binary.

    $ go install

Installs the **godist** binary to your *GOPATH*/bin.

    $ go run main.go

Useful during debug.

## Contributing

Pull requests, issues, and comments are greatly appreciated.

## License

Godist is licensed under the MIT license, see the LICENSE file.