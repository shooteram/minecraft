windows:
	make windows_amd64
	make windows_386

windows_amd64:
	CC=i586-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
		go build -v -o "builds/minecraft_launcher_cmd_windows_amd64.exe" -ldflags="-extld=$CC"

windows_386:
	CC=i586-mingw32-gcc GOOS=windows GOARCH=386 CGO_ENABLED=1 \
		go build -v -o "builds/minecraft_launcher_cmd_windows_386.exe" -ldflags="-extld=$CC"

linux:
	make linux_amd64
	make linux_386

linux_amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -v -o "builds/minecraft_launcher_cmd_linux_amd64"
linux_386:
	GOOS=linux GOARCH=386 go build -v -o "builds/minecraft_launcher_cmd_linux_386"

darwin:
	make darwin_amd64
	make darwin_386

darwin_amd64:
	GOOS=darwin GOARCH=amd64 go build -v -o "builds/minecraft_launcher_cmd_darwin_amd64"
darwin_386:
	GOOS=darwin GOARCH=386 go build -v -o "builds/minecraft_launcher_cmd_darwin_386"

amd64:
	make windows_amd64
	make linux_amd64
	make darwin_amd64

386:
	make windows_386
	make linux_386
	make darwin_386

all:
	make linux
	make windows
	make darwin
