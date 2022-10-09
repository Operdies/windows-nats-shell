package main

import (
	"testing"

	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"gopkg.in/yaml.v3"
)

func TestGetCustomConfig(t *testing.T) {
	var testString = []byte(`
custom:
  launcher:
    includesystempath: true
    watchsystempath: true
    sources:
      - path: C:/Users/alexw/scripts
        recursive: true 
        watch: false
      - path: "C:/Users/alexw/AppData/Roaming/Microsoft/Windows/Start Menu/"
        recursive: true 
        watch: false
      - path: "C:/ProgramData/Microsoft/Windows/Start Menu/"
        recursive: false 
        watch: true
executable: ./bin/windows-nats-driver.exe
workingdirectory: C:/Users/alexw/repos/minimalist-shell
autorestart: true
`)
	var cfg shell.Service
	yaml.Unmarshal(testString, &cfg)
	result, err := shell.GetCustom[customOptions](cfg)
	if err != nil {
		t.Fatalf("GetCustom failed: %v", err)
	}
	if len(result.Launcher.Sources) != 3 {
		t.Fatalf("Failed parsing config. Got:\n%v\nFrom:\n%v", result, string(testString))
	}
	if result.Launcher.IncludeSystemPath == false {
		t.Fatal("Expected IncludeSystemPath=true")
	}
	scripts := result.Launcher.Sources[0]
	if scripts.Recursive != true || scripts.Watch != false {
		t.Fatal("Parse error in scripts")
	}
	roaming := result.Launcher.Sources[1]
	if roaming.Recursive != true || roaming.Watch != false {
		t.Fatal("Parse error in roaming")
	}
	programdata := result.Launcher.Sources[2]
	if programdata.Recursive != false || programdata.Watch != true {
		t.Fatal("Parse error in programdata")
	}
}
