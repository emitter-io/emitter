FROM microsoft/dotnet:1.1-sdk
MAINTAINER Roman Atachiants "roman@misakai.com"

# Make sure we have S3 & additional libraries
RUN apt-get update -qq \
	&& apt-get install -y git \
	&& mkdir /tmp/emitter \
	&& cd /tmp/emitter \
	&& git clone "https://github.com/emitter-io/emitter.git" "/tmp/emitter" \
	&& cd /tmp/emitter \
	&& dotnet restore -v Minimal \
	&& cd /tmp/emitter/src/Emitter.Server \
	&& dotnet publish -c Release --output /app \
	&& rm -rf /tmp/emitter

# Application will be in app folder
WORKDIR /app
ADD deploy.sh /

# Emitter ports
EXPOSE 8080
EXPOSE 4000

# Volumes for faster deploy
VOLUME /app
VOLUME /root/.nuget/packages
CMD ["/bin/bash", "/deploy.sh"]