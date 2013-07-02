# Godist
**– _Transmission control protocol (TCP) distribution_**

Godist distributes your incoming tcp connections between servers, balancing the load. A thin high-performance layer between your servers and the world.

Written in Go. Built for concurrency.

## Usage
Godist requires a configuration file to setup the most basic settings. The software will look for **godist.conf** by default. You can specify your own configuration file by passing it as an argument.

**godist.conf**

    {
		"Host": "localhost:8080",
		"Sockets": 
			["localhost:4000", 
			 "localhost:4001",
			 "localhost:4002"]
	}

Then:

    $ godist godist.conf

### Distributing connections
The incoming connections are distributed by hashing the ip and port. An ip and port will always get the same server handling its connection.

## Contributing

Pull requests, issues, and comments are greatly appreciated.

## License

Godist is licensed under the MIT license, see the LICENSE file.