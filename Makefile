BINPATH=./bin

all: 
	go build -o $(BINPATH) ./...

hookdll: 
		gcc -O3 -shared  -I./c/hook-dll  -fpic ./c/hook-dll/hook.c -o $(BINPATH)/libhook.dll

test:
	go test -v ./...

events: 
	go build -ldflags -H=windowsgui -o $(BINPATH)/events.exe ./cmd/events/

driver: 
	go build -o $(BINPATH)/driver.exe ./cmd/driver/

shell: 
	go build -o $(BINPATH)/shell.exe ./cmd/shell/

background: 
	go build -o $(BINPATH)/background.exe ./cmd/background/

hotkeys: 
	go build -o $(BINPATH)/hotkeys.exe ./cmd/hotkeys/ && nats pub Shell.RestartService hotkeys
