# snap publisher plugin - blueflood

[![Build Status](https://travis-ci.org/Staples-Inc/snap-plugin-publisher-blueflood.svg?branch=master)](https://travis-ci.org/Staples-Inc/snap-plugin-publisher-blueflood)

Publishes data to [Blueflood](http://blueflood.io/), a [Cassandra](http://cassandra.apache.org/) based ingest engine.

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Operating Systems](#operating-systems))
  * [Build](#build)
2. [Contributing](#contributing)
3. [License](#license)

## Getting Started
A working snap agent and a running instance of Blueflood is required to use this plugin.

### System Requirements
* [golang 1.5+](https://golang.org/dl/)
* [snap](https://github.com/intelsdi-x/snap)
* [blueflood](http://blueflood.io/)
* [cassandra](http://cassandra.apache.org/)

### Operating System
* Linux
* Mac OS X

### Build
Fork https://github.com/Staples-Inc/snap-plugin-publisher-blueflood
Clone repo into `$GOPATH/src/github.com/Staples-Inc/`:

```
$ git clone https://github.com/<yourGithubID>/snap-plugin-publisher-blueflood.git
```

Build the plugin by running make within the cloned repo:
```
$ make
```
This builds the plugin in `/build/rootfs/`

## Contributing
We currently have no future plans for this plugin. If you have a feature request, please add it as an issue and/or submit a pull request

## License
This plugin is Open Source software released uder the Apache 2.0 [License](LICENSE)
