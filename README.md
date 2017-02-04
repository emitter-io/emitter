![](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/logo/emitter_logo_blue.png)

# Emitter: Clustered Publish-Subscribe Broker
> [Emitter is a free open source real-time messaging service](https://emitter.io) that connects all devices. This publish-subscribe messaging API is built for speed and security.

Emitter is a real-time communication service for connecting online devices. Infrastructure and APIs for IoT, gaming, apps and real-time web. At its core, emitter.io is a distributed, scalable and fault-tolerant publish-subscribe messaging platform based on MQTT protocol and featuring message storage.

Emitter can be used for online gaming and mobile apps by satisfying the requirements for low latency, binary messaging and high throughput. It can also be used for the real-time web application such as dashboards or visual analytics or chat systems. Moreover, Emitter is perfect for the internet of things and allows sensors to be controlled and data gathered and analyzed.

[![Join the chat at https://gitter.im/emitter-io/public](https://badges.gitter.im/emitter-io/public.svg)](https://gitter.im/emitter-io/public?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) 
[![Build status](https://ci.appveyor.com/api/projects/status/6im4291ao9i664ix?svg=true)](https://ci.appveyor.com/project/Kelindar/emitter)
[![Twitter Follow](https://img.shields.io/twitter/follow/emitter_io.svg?style=social&label=Follow)](https://twitter.com/emitter_io)

## Server Quick Start

The quick way to start an Emitter broker is by using `docker run` command as shown below. Alternatively, you might compile this repository and use `dotnet` CLI or Visual Studio to rebuild and run from source. 

```shell
docker run -d --name emitter -p 8080:8080 --privileged --restart=unless-stopped emitter/server
```
The command above starts a new server and if no configuration or environment variables were supplied, it will print out a message similar to the message below once the server has started:

```shell
Warning: New license: BjeUWk46tfTTL6ks5q-Vnyj-puoAAAAAAAAAAAAAAAI
Warning: New secret key: Hc4pyBAGEe6Z9PYy77AH0Y43dQm62faH
...
Listening: 127.0.0.1:8080
Listening: 127.0.0.1:8443
Listening: 127.0.0.1:4000
```

This message shows that a new security configuration was generated, you can then re-run EMITTER_LICENSE set to the specified value. Alternatively, you can set `"license"` property in the `emitter.conf` configuration file.

Finally, open a browser and navigate to **<http://127.0.0.1:8080/keygen>** in order to generate your key. Now you can use the secret key generated to create channel keys, which allow you to secure individual channels and start using emitter.

## Sandbox

Emitter has a [sandbox](https://emitter.io/login) - a free cloud cluster which allows you to quickly try out the platform and see how simple it is to create connected, real-time applications. The movie below shows you how to create your sandbox account and create a simple hello-world application within **5 minutes**.

[![Getting Started With Emitter](https://s3-eu-west-1.amazonaws.com/emitter.io/content/img/wiki/thumb1.png)](https://www.youtube.com/watch?v=WyPMeIgfxSM "Getting Started With Emitter")


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

## Building

The server requires [.NET Core](https://www.microsoft.com/net/core) platform to be installed. Once you have this installed, simply clone this repository and run the following commands to restore the Nuget packages and run the server.

```shell
dotnet restore
cd src/Emitter.Server
dotnet run
```
Alternatively, you can use [Visual Studio IDE](https://www.visualstudio.com/) to build, run and debug. Simply open the `Emitter.sln` file provided.

## Deploying as Docker Container

[![Docker Automated build](https://img.shields.io/docker/automated/emitter/server.svg)](https://hub.docker.com/r/emitter/server/)
[![Docker Pulls](https://img.shields.io/docker/pulls/emitter/server.svg)](https://hub.docker.com/r/emitter/server/)

Emitter is convinently packaged as a docker container. To run the emitter service on a single server, use the command below. Once the server is started, it will generate a new security configuration, you can then re-run the same command with an additional environment variable -e EMITTER_LICENSE set to the provided value.

```shell
docker run -d -p 8080:8080 emitter/server
```
For the clustered (multi-server) mode, the container can be started using the simple docker run with 3 main parameters.

```shell
docker run -d -p 8080:8080 -p 4000:4000 -e EMITTER_LICENSE=[key] -e EMITTER_CLUSTER_SEED=[seed] -e EMITTER_CLUSTER_KEY=[name] emitter/server
```

## Support, Discussion, and Community

If you need any help with Emitter Server or any of our client SDKs, please join us at either our [gitter chat](https://gitter.im/emitter-io/public) where most of our team hangs out at or drop us an e-mail at <info@emitter.io>.

Please submit any Emitter bugs, issues, and feature requests to emitter-io>emitter. If there are any security issues, please email info@emitter.io instead of posting an open issue in Github.


## Contributing

If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are warmly welcome.

## Licensing

Copyright (c) 2009-2016 Misakai Ltd. This project is licensed under [Affero General Public License v3](https://github.com/emitter-io/emitter/blob/master/LICENSE).
