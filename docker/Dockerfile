FROM misakai/mono:4.2.1
MAINTAINER Roman Atachiants "roman@emitter.io"

# Make sure we have additional libraries
RUN apt-get update -qq \
	&& apt-get install -y libgdiplus libcurl3 libxml2 wget libc6 libc6-dev

# Application will be in app folder
WORKDIR /app
ADD . /app

# Emitter ports
EXPOSE 8080
EXPOSE 4000

CMD ["/bin/bash", "/app/deploy.sh"]