go test ./...
%GOPATH%/bin/overalls -project=github.com\emitter-io\emitter
%GOPATH%/bin/goveralls -coverprofile=overalls.coverprofile -service=appveyor-ci -repotoken=%COVERALLS_TOKEN%

call :build linux arm  
call :build linux amd64  
call :build linux 386  
call :build darwin amd64  
call :build darwin 386   
call :build windows 386 .exe
call :build windows amd64 .exe
goto :end

:build
	set GOOS=%1
	set GOARCH=%2
	go tool dist install pkg/runtime
	go install -a std
	go build -o build/emitter-%1-%2%3 -i  -ldflags "-X main.emitterVersion=%APPVEYOR_BUILD_VERSION% main.emitterCommit=%APPVEYOR_REPO_COMMIT%" .
:end