call :build linux arm  
call :build linux amd64  
call :build linux 386  
call :build darwin amd64  
call :build darwin 386   
call :build windows amd64 .exe
call :build windows 386 .exe
goto :end

:build
	set GOOS=%1
	set GOARCH=%2
	go tool dist install pkg/runtime
	go install -a std
	go build -o build/emitter-%1-%2%3 -i .
:end