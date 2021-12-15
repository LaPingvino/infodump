# Infodump

Infodump is a commandline tool that serves as a social network, initially for the neurodivergent community. It is peer-to-peer, and is based on the [IPFS](https://ipfs.io) distributed file system and written in [Go](https://golang.org/). It is a simple, easy-to-use, and open-source project. It is also an excuse for me to use Github's Copilot feature to write code and documentation, and a means to test and expand my OLN ideas: creating a network that enables topic and location-based communication.

This is accomplished through the PubSub functionality of IPFS, as well as a local database that stores all the information until it is synced up with other peers. It also uses a Hashcash-based proof of work system to enable an ephemeral approach to the network: over time messages will be removed from the network if they are not specifically saved.

Another feature of Infodump is the ability to create and save topics and locations (almost nothing of that is implemented yet). These are used to create an ad-hoc network of people who share a common interest, without the need for user accounts and authentication.

Even though the code is in a VERY early stage, I encourage you to try it out and maybe even contribute to it.
## Getting started

You need to have a copy of IPFS and the Go compiler installed, and you need to run IPFS with the `--enable-pubsub-experiment` flag.

You can get IPFS from [the IPFS website](https://ipfs.io/) and Go from [the Go website](https://golang.org/). If you haven't already, prepare IPFS by running `ipfs init` and then start the daemon with `ipfs daemon --enable-pubsub-experiment`.

To install Infodump, run `go install git.kiefte.eu/lapingvino/infodump@latest` while making sure that the Go bin directory is in your PATH in order to compile the binary and run it.

This is very experimental software, and I am not responsible for any damage that may be caused by using it. Use at your own risk. Please report any bugs you find. I will also be very happy with any code contributions and even forks. I am especially interested in nice looking web GUIs to the network; if you create a proof of concept of such, you are my hero.