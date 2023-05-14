# dnsreverseproxy
## A very basic DNS reverse proxy 

I started this project because I had a particular itch. 
I wanted to create a middle DNS that could rewrite my home domain into Consul DNS queries while also routing the 
remaining queries to a regular DNS. 
Can consul manage a different domain other than .consul? Of course, but I wanted to get rid of the .service part so
I have something like gitlab.example.com that goes to the same IP as gitlab.service.consul. 
I also played with prepared queries, but the documentation I've found online was not enough for me and I ended up
doing this directly so I don't have a very complicated setup.
dnsdist is another candidate, but I preferred to do it in golang to learn the language.

## Usage
```
COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --listen value, -l value                                                       Address and port to listen on (default: ":1053")
   --masqued-domain value, -m value                                               Domain to be masqued (default: ".example.com.")
   --upstream-domain value, -u value                                              Upstream domain to be masqued (default: ".service.consul.")
   --upstream-special value, --uc value [ --upstream-special value, --uc value ]  Special upstream host:port (default: "localhost:8600")
   --upstream-normal value, --un value [ --upstream-normal value, --un value ]    Normal resolution upstream host:port (default: "1.1.1.1:53")
   --help, -h                                                                     show help
```
