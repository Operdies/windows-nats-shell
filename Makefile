BINPATH=./bin
BUILDFLAGS=

all: events driver shell hotkeys windowmanager toast background # very slow

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

hotkeys: 
	go build -o $(BINPATH)/hotkeys.exe ./cmd/hotkeys/ && nats pub Shell.RestartService hotkeys

windowmanager: 
	go build -o $(BINPATH)/windowmanager.exe ./cmd/windowmanager/ && nats pub Shell.RestartService windowmanager

toast: 
	go build -o $(BINPATH)/toast.exe ./cmd/toast/ && nats pub Shell.RestartService toast

background: 
	go build -o $(BINPATH)/background.exe ./cmd/background/ && nats pub Shell.RestartService background
