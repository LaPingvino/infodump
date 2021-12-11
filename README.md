# Infodump

Infodump is a commandline tool that serves as a social network, initially for the neurodivergent community. It is peer-to-peer, and is based on the [IPFS](https://ipfs.io) distributed file system and written in [Go](https://golang.org/). It is a simple, easy-to-use, and open-source project. It is also an excuse for me to use Github's Copilot feature to write code and documentation, and a means to test and expand my OLN ideas: creating a network that enables topic and location-based communication.

This is accomplished through the PubSub functionality of IPFS, as well as a local database that stores all the information until it is synced up with other peers. It also uses a Hashcash-based proof of work system to enable an ephemeral approach to the network: over time messages will be removed from the network if they are not specifically saved.

Another feature of Infodump is the ability to create and save topics and locations (almost nothing of that is implemented yet). These are used to create an ad-hoc network of people who share a common interest, without the need for user accounts and authentication.

Even though the code is in a VERY early stage, I encourage you to try it out and maybe even contribute to it. I am especially interested in nice looking web GUIs to the network; if you create a proof of concept of such, you are my hero.