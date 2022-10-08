BINPATH=./bin

all: hookdll
	go build -o $(BINPATH) ./...

test:
	go test -v ./...

hookdll: 
	gcc -O3 -shared  -I./c/hook-dll  -fpic ./c/hook-dll/hook.c -o $(BINPATH)/libhook.dll

shell-event-publisher: 
	go build -o $(BINPATH)/shell-event-publisher.exe ./cmd/shell-event-publisher/

windows-nats-driver: 
	go build -o $(BINPATH)/windows-nats-driver.exe ./cmd/windows-nats-driver/

windows-shell: 
	go build -o $(BINPATH)/windows-shell.exe ./cmd/windows-shell/

background: 
	go build -o $(BINPATH)/background.exe ./cmd/background/
