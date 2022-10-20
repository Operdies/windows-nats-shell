# diary

Scribblings of a madman

## Create rofi integration

Investigate what is possible for rofi modules
rofi -dmenu only supports text. This can be used to support many cases,
but makes scripting difficult if the output should match partial strings, e.g. the hwnd from a 'SetFocus' request
With real integration, the launcher can also ship real icons

## GetProcessList nats subject

GetProcessList should contain extenteded, useful information
about the process. GetWindows is probably still different enough
that they can both coexist.
Evaluate if NewProcess events are necessary

## Launcher process IO

The launcher could also support IO streams. This should probably be a different subject though.
StartProcess instead of LaunchProgram? Reading stdout/err should behave like streams here though, and not be published.
StartProcessElevated can be implemented by launching a go process elevated and hosting it there.
Update: this almost sounds like a service. Support for this should probably be something like `Shell.AddService(yaml blob)`.

## NATS log service

logs for services could be stored in a sqlite database
and be queried using NATS. Then there would be no need
for a console window to host the shell
Since logs are now published in nats, this database could just record everything

## Kill menu

Make an API to kill a process by its handle

## Background

API for posting data (images or text) to the background. TBD: pre-configured zones or defined per request / client?

## Window manager

Some (toggle-able) way to automatically tile windows. BSPC inspired layouts
Cycle between revolver strategy, tall mode, and monocle mode -- revolver with max size is monocle..
Disable borders on all windows (maybe with hotkey to enable currently focused window?)

* add toggle for auto-layout
* swap main / secondary windows when a window is manually switched (alt-tab or click)
  - Respect manual order overrides when manually swapping
* Highlight border of focused window 
* Avoid removing borders from windows that don't have a border in the first place (chrome, steam) -- figure out how to detect this?

* Make abstractions more sane -- inputhandler duplicates hotkey functionality
* a lot of `cmd/*` code is generic and could be moved to a suitable package


## Steam integration

Launch / install games
Investigate what integrations exist / are possible

## Rofi but for windows

Would make the shell usable without a linux driver. The shortcut manager needs to support input/output. Then the rofi implementation can respond using nats

## Bugs 

* Processes created with `driver`'s `System.LaunchProgram` will run as admin if `driver` is running as admin.
- Consider if we need a full-blown CreateProcess implementation which mines the registry and properly controls inherited handles
* `driver` cannot open e.g. a `.png` file after we switched to `ShellExecute`.
- Add handlers to config?

## Thought cabinet

> Service namespaces?
> Services where multiple instances make sense should really use namespaces. The clients should probably support namespaces in some way
> The environment variable containing the service name could just be prepended as a subject where it makes sense. But then
> then e.g. Requester clients also have to require the namespace as an input. Do clients care which instance responds?

> How can input/output be implemented? Named actions?

`... payload: { hwnd: $action1.hwnd, command: $action2.command }` ?
Then actions with no dependencies can simply be `publish`ed. A dependency tree must be built of other actions.
There should probably be a panic during startup if circular dependencies are detected.
There should also be a startup panic if collisions are detected.

> Should all APIs require complex type inputs ?

Should all e.g. Windows APIs require a window as input? If an action only requires e.g. `hwnd` to be set, the input can just be `hwnd: 1234`
More generally, should all X apis take an input of the same form as their outputs? It would make it possible to return meaningful errors.
Rofi integration is a prerequisite for this, because the alternative is way too complicated on the scripting side.

> Current implementation is extremely windows 

We don't need full windows support, but it would be nice if the project was structured such that platform dependent parts were neatly separated 
so the project can at least compile
