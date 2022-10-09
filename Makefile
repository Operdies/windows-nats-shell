BINPATH=./bin

all: hookdll
	go build -o $(BINPATH) ./...

test:
	go test -v ./...

hookdll: 
	gcc -O3 -shared  -I./c/hook-dll  -fpic ./c/hook-dll/hook.c -o $(BINPATH)/libhook.dll

events: 
	go build -o $(BINPATH)/events.exe ./cmd/events/

driver: 
	go build -o $(BINPATH)/driver.exe ./cmd/driver/

shell: 
	go build -o $(BINPATH)/shell.exe ./cmd/shell/

background: 
	go build -o $(BINPATH)/background.exe ./cmd/background/

hotkeys: 
	go build -o $(BINPATH)/hotkeys.exe ./cmd/hotkeys/
