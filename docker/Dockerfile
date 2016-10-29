FROM microsoft/dotnet:1.0.0-preview2-sdk
MAINTAINER Roman Atachiants "roman@misakai.com"

# Make sure we have S3 & additional libraries
RUN apt-get update -qq \
	&& apt-get install -y s3cmd git wget \
	&& git clone "https://github.com/emitter-io/emitter.git" "/app" \
	&& cd /app \
	&& dotnet restore -v Minimal

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