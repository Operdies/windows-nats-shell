BINPATH=./bin

all: events driver shell background hotkeys

hookdll: 
		gcc -O3 -shared  -I./c/hook-dll  -fpic ./c/hook-dll/hook.c -o $(BINPATH)/libhook.dll

test:
	go test -v ./...

events: 
	go build -o $(BINPATH)/events.exe ./cmd/events/ && nats pub Shell.RestartService events

driver: 
	go build -o $(BINPATH)/driver.exe ./cmd/driver/ && nats pub Shell.RestartService driver

shell: 
	go build -o $(BINPATH)/shell.exe ./cmd/shell/

background: 
	go build -o $(BINPATH)/background.exe ./cmd/background/ && nats pub Shell.RestartService background

hotkeys: 
	go build -o $(BINPATH)/hotkeys.exe ./cmd/hotkeys/ && nats pub Shell.RestartService hotkeys
