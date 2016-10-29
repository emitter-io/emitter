#!/bin/bash 
# Defaults
: ${APP_FILE='emitter-service.zip'} 

# Download the zip
wget -O emitter-service.zip http://app.misakai.com.s3.amazonaws.com/public/Emitter/bin/$APP_FILE

# Get the application archive
if [[ -z $APP_ARCHIVE ]]; then
	APP_ARCHIVE=$(find /app -name "*.zip")
fi

# Make sure we have certificates we need
mozroots --import --machine --sync

# Unzip the package and delete the zip file
unzip -qq -o $APP_ARCHIVE -d /app
rm $APP_ARCHIVE

# Optionally enable ahead-of-time compilation
if [[ $MONO_ENABLE_AOT ]]; then
	mono --aot -O=all ${APP_ENTRY}
fi

# Use specific garbage collector
if [[ -z  $MONO_USE_GC ]]; then
	MONO_USE_GC=sgen
fi

# Set maximum threads per CPU 
if [[ -z $MONO_THREADS_PER_CPU ]]; then
	MONO_THREADS_PER_CPU=100
fi

# Run the application
mono --server --gc=$MONO_USE_GC Emitter.exe