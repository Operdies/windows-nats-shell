BINPATH=./bin

test:
	go test -v ./...

libevents: 
	go build -buildmode=c-shared -o $(BINPATH)/libevents.dll ./cmd/events/libevents/

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

all: events driver shell background hotkeys libevents 
