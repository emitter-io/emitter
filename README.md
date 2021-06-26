<p align="center">
<img src="https://s3.amazonaws.com/cdn.misakai.com/www-emitter/logo/emitter_logo_blue.png" border="0" alt="kelindar/column">
<br>
<img src="https://img.shields.io/github/go-mod/go-version/emitter-io/emitter" alt="Go Version">
<a href="https://goreportcard.com/report/github.com/emitter-io/emitter"><img src="https://goreportcard.com/badge/github.com/emitter-io/emitter" alt="Go Report Card"></a>
<a href="https://coveralls.io/github/emitter-io/emitter"><img src="https://coveralls.io/repos/github/emitter-io/emitter/badge.svg" alt="Coverage"></a>
<a href="https://twitter.com/emitter_io"><img src="https://img.shields.io/twitter/follow/emitter_io.svg?style=social&label=Follow" alt="Twitter"></a>
</p>

## Emitter: Distributed Publish-Subscribe Platform

Emitter is a distributed, scalable and fault-tolerant publish-subscribe  platform built with MQTT protocol and featuring message storage, security, monitoring and more:
* **Publish/Subscribe** using MQTT over TCP or Websockets.
* Resilient, **highly available and partition tolerant** (AP in CAP terms).
* Able to handle **3+ million of messages/sec** on a single broker.
* Supports **message storage** with history and **message-level expiry**.
* Provides **secure channel keys** with permissions and can face the internet.
* Automatic **TLS/SSL** and encrypted inter-broker communication.
* Built-in **monitoring with Prometheus, StatsD** and more.
* **Shared subscriptions, links** and private links for channels.
* Easy **deployment with Docker and Kubernetes** of production-ready clusters.

Emitter can be used for online gaming and mobile apps by satisfying the requirements for low latency, binary messaging and high throughput. It can also be used for the real-time web application such as dashboards or visual analytics or chat systems. Moreover, Emitter is perfect for the internet of things and allows sensors to be controlled and data gathered and analyzed.

## Tutorials & Demos

The following [video tutorials](https://www.youtube.com/playlist?list=PLhFXrq-2gEb0ygxR477GJLngjYu-FcSVq) demonstrate various features of Emitter in action.

[![FOSDEM 2018](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-fosdem2018.png)](https://www.youtube.com/watch?v=M8VhWckhZoM)
[![FOSDEM 2019](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-fosdem2019.png)](https://www.youtube.com/watch?v=GZDgN8XHy7g)
[![PubSub in Go](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-golang.png)](https://www.youtube.com/watch?v=ggFqj5P4W38)
[![Message Storage](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-storage.png)](https://www.youtube.com/watch?v=14cIxnR0Akc)
[![Using MQTTSpy](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-mqttspy.png)](https://www.youtube.com/watch?v=OcdL_454XT0)
[![ISS Tracking](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-iss.png)](https://www.youtube.com/watch?v=F47LTbl2Bjw)
[![Self-Signed TLS](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-tls.png)](https://www.youtube.com/watch?v=NRejmavZuAI)
[![Monitor with eTop](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-etop.png)](https://www.youtube.com/watch?v=EOlOk9JPSyA)
[![StatsD and DataDog](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-datadog.png)](https://www.youtube.com/watch?v=bi77Wb7cqEc)
[![Links & Private Links](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-links.png)](https://www.youtube.com/watch?v=_FgKiUlEb_s)
[![Building a Client-Server app with Publish-Subscribe in Go](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-golang.png)](https://www.youtube.com/watch?v=SHPkJ6ESnpw)
[![Distributed Actor Model with Publish/Subscribe and Golang](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-actor.png)](https://www.youtube.com/watch?v=nWPKlopWOss)
[![Online Multiplayer Platformer Game with Emitter](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-platformer.png)](https://www.youtube.com/watch?v=UIzMTyfLq7Y)
[![Keeping one Last Message per Channel using MQTT Retain](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-retain.png)](https://www.youtube.com/watch?v=w3BXfYqYAvg)
[![Load-balance Messages using Subscriber Groups](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-share.png)](https://www.youtube.com/watch?v=Vl7iGKEQrTg)

## How to Deploy
[![Local Emitter Cluster](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-win.png)](https://www.youtube.com/watch?v=byq70fHeH-I)
[![K8s and DigitalOcean](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-k8s.png)](https://www.youtube.com/watch?v=CsrKiNjZ2Ew)
[![K8s and Google Cloud](http://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-gcloud.png)](https://www.youtube.com/watch?v=IL7WEH_2IOo)
[![K8s and Azure](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-azure.png)](https://www.youtube.com/watch?v=4ixnxreKsOg)

## Quick Start

### Run Server
The quick way to start an Emitter broker is by using `docker run` command as shown below. 

Notice: You must use -e for docker environment.You could get it from your docker log.

```shell
docker run -d --name emitter -p 8080:8080 --restart=unless-stopped emitter/server
```

Alternatively, you might compile this repository and use `go get` command to rebuild and run from source. 

```shell
go get -u github.com/emitter-io/emitter && emitter
```

### Get License
Both commands above start a new server and if no configuration or environment variables were supplied, it will print out a message(or you could find it in the docker log) similar to the message below once the server has started :

```shell
[service] unable to find a license, make sure 'license' value is set in the config file or EMITTER_LICENSE environment variable
[service] generated new license: uppD0PFIcNK6VY-7PTo7uWH8EobaOGgRAAAAAAAAAAI
[service] generated new secret key: JUoOxjoXLc4muSxXynOpTc60nWtwUI3o
```

This message shows that a new security configuration was generated, you can then re-run `EMITTER_LICENSE` set to the specified value. Alternatively, you can set `"license"` property in the `emitter.conf` configuration file.

### Re-Run Command
References are available. Please replace your EMITTER_LICENSE to it.
```shell
docker run -d --name emitter -p 8080:8080 -e EMITTER_LICENSE=uppD0PFIcNK6VY-7PTo7uWH8EobaOGgRAAAAAAAAAAI --restart=unless-stopped emitter/server
```

### Generate Key
Finally, open a browser and navigate to **<http://127.0.0.1:8080/keygen>** in order to generate your key. Now you can use the secret key generated to create channel keys, which allow you to secure individual channels and start using emitter.

**Warning:** If you use upon command, you secret is JUoOxjoXLc4muSxXynOpTc60nWtwUI3o. And it's not safe!!!


## Usage Example

The code below shows a small example of usage of emitter with the Javascript SDK. As you can see, the API exposes straightforward methods such as `publish` and `subscribe` which can take binary payload and are secured through channel keys.

```javascript
// connect to emitter service
var connection = emitter.connect({ host: '127.0.0.1' });

// once we're connected, subscribe to the 'chat' channel
emitter.on('connect', function(){
    emitter.subscribe({
        key: "<channel key>",
        channel: "chat"
    });
});

// publish a message to the chat channel
emitter.publish({
    key: "<channel key>",
    channel: "chat/my_name",
    message: "hello, emitter!"
});
```

Further documentation, demos and language/platform SDKs are available in the [**develop section of our website**](https://emitter.io/develop). Make sure to check out the [**getting started tutorial**](https://emitter.io/develop/getting-started) which explains the basic usage of emitter and MQTT.

## Command line arguments

The Emitter broker accepts command line arguments, allowing you to specify a configuration file, usage is shown below.

```shell
-config string
   The configuration file to use for the broker. (default "emitter.conf")

-help
   Shows the help and usage instead of running the broker.
```

## Configuration File

The configuration file (defaulting to `emitter.conf`) is the main way of configuring the broker. The configuration file is however, not the only way of configuring it as it allows a multi-level override through **environment variables** and/or  **hashicorp Vault**. 

The configuration file is in JSON format, but you can override any value by providing an environment variable which follows a particular format, for example if you'd  like to provide a `license` through environment variable, simply define `EMITTER_LICENSE` environment variable, similarly, if you want to specify a `certificate`, define `EMITTER_TLS_CERTIFICATE` environment variable. Example of configuration file:

```json
{
    "listen": ":8080",
    "license": "/*The license*/",
    "tls": {
        "listen": ":443",
        "host": "example.com"
    },
    "cluster": {
        "listen": ":4000",
        "seed": " 192.168.0.2:4000",
        "advertise": "public:4000"
    },
    "storage": {
        "provider": "inmemory"
    }
}
```

The structure of the configuration is described below:

| Property | Env. Variable | Description |
|---|---|---|
| `license` | `EMITTER_LICENSE` | The license file to use for the broker. This contains the encryption key. |
| `listen` | `EMITTER_LISTEN` | The API address used for TCP & Websocket communication, in `IP:PORT` format (e.g: `:8080`). |
| `limit.messageSize` | `EMITTER_LIMIT_MESSAGESIZE` | Maximum message size. Default is 64KB.
| `tls.listen` | `EMITTER_TLS_LISTEN` |The API address used for Secure TCP & Websocket communication, in `IP:PORT` format (e.g: `:443`).  |
| `tls.host` | `EMITTER_TLS_HOST` | The hostname to whitelist for the certificate.  |
| `tls.email` | `EMITTER_TLS_EMAIL` |The email account to use for autocert. |
| `vault.address` | `EMITTER_VAULT_ADDRESS` | The Hashicorp Vault address to use to further override configuration. |
| `vault.app` | `EMITTER_VAULT_APP` | The Hashicorp Vault application ID to use. |
| `cluster.name` | `EMITTER_CLUSTER_NAME` | The name of this node. This must be unique in the cluster. If this is not set, Emitter will set it to the external IP address of the running machine. |
| `cluster.listen` | `EMITTER_CLUSTER_LISTEN` | The IP address and port that is used to bind the inter-node communication network. This is used for the actual binding of the port. |
| `cluster.advertise` | `EMITTER_CLUSTER_ADVERTISE` | The address and port to advertise inter-node communication network. This is used for nat traversal. |
| `cluster.seed` | `EMITTER_CLUSTER_SEED` | The seed address (or a domain name) for cluster join. |
| `cluster.passphrase` | `EMITTER_CLUSTER_PASSPHRASE` | Passphrase is used to initialize the primary encryption key in a keyring. This key is used for encrypting all the gossip messages (message-level encryption). |
| `storage.provider` | `EMITTER_STORAGE_PROVIDER` |  This property represents the publishers publish message storage mode. there are two kinds of can use, they are respectively `inmemory` and `ssd`, defaults to the former. |
| `storage.config.dir` | `EMITTER_STORAGE_CONFIG` |  If the storage mode is `ssd`, this property indicates where the messages are stored (emitter server nodes are not allowed to use the same directory within the same machine)



## Building and Testing

The server requires [Golang 1.9](https://golang.org/dl/) to be installed. Once you have this installed, simply `go get` this repository and run the following commands to download the package and run the server.

```shell
go get -u github.com/emitter-io/emitter && emitter
```

If you want to run the tests, simply run `go test` command as demonstrated below.

```shell
go test ./...
```

## Deploying as Docker Container

[![Docker Automated build](https://img.shields.io/docker/automated/emitter/server.svg)](https://hub.docker.com/r/emitter/server/)
[![Docker Pulls](https://img.shields.io/docker/pulls/emitter/server.svg)](https://hub.docker.com/r/emitter/server/)

Emitter is conveniently packaged as a docker container. To run the emitter service on a single server, use the command below. Once the server is started, it will generate a new security configuration, you can then re-run the same command with an additional environment variable -e EMITTER_LICENSE set to the provided value.

```shell
docker run -d -p 8080:8080 emitter/server
```
For the clustered (multi-server) mode, the container can be started using the simple docker run with 3 main parameters.

```shell
docker run -d -p 8080:8080 -p 4000:4000 -e EMITTER_LICENSE=[key] -e EMITTER_CLUSTER_SEED=[seed] -e EMITTER_CLUSTER_PASSPHRASE=[name] emitter/server
```

## Support, Discussion, and Community

[![Join the chat at https://gitter.im/emitter-io/public](https://badges.gitter.im/emitter-io/public.svg)](https://gitter.im/emitter-io/public?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) 

If you need any help with Emitter Server or any of our client SDKs, please join us at either our [gitter chat](https://gitter.im/emitter-io/public) where most of our team hangs out at or drop us an e-mail at <info@emitter.io>.

Please submit any Emitter bugs, issues, and feature requests to emitter-io>emitter. If there are any security issues, please email info@emitter.io instead of posting an open issue in Github.


## Contributing

If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are warmly welcome.

## Licensing

Copyright (c) 2009-2019 Misakai Ltd. This project is licensed under [Affero General Public License v3](https://github.com/emitter-io/emitter/blob/master/LICENSE).

Emitter offers support contracts and is now also offered via a commercial license. Please contact info@emitter.io for more information.
