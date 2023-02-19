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
Should this API support different segments to have different Z-orders (even drawing on top of windows?)

## Notification system 

This goes hand in hand with posting to the background. In general, we really need a text shader.
A text shader will also enable launchers etc

## Window manager

Some (toggle-able) way to automatically tile windows. BSPC inspired layouts
Cycle between revolver strategy, tall mode, and monocle mode -- revolver with max size is monocle..

* add toggle for auto-layout
* Highlight border of focused window 
* glfw 3.4 adds click-through windows, but it might not be released for a while, and is not available in go-glfw
* add something analogous to hiding and restoring windows 
- Currently, hidden windows must be restored by alt-tabbing, which means alt-tab cannot be overridden.

* Make abstractions more sane -- inputhandler duplicates hotkey functionality
the hotkey manager should be easier to use. It's not really tennable to require a nats endpoint 
in order to add a shortcut for something. Investigate what can be done with reflection.
-- Shortcuts is inherently something that is running locally on the machine which the keyboard is connected to. 
* a lot of `cmd/*` code is generic and could be moved to a suitable package

## Windows app launcher 
* Make Rofi but for windows already. Generic pickers is a huge no-brainer.

## Steam integration

Launch / install games
Investigate what integrations exist / are possible

## Rofi but for windows

Would make the shell usable without a linux driver. The shortcut manager needs to support input/output. Then the rofi implementation can respond using nats

## Status bar 
Now that we have pseudo-tiling and padding, we can start working on a status bar. 
The taskbar should probably be drawn on top of windows because the alternative would mess with Revolver

## System Tray

I don't really like trays but I guess I need them. Place them on background?
-- Tray icons may not even be possible without explorer? It seems they are very tightly integrated with 
-- exlorer.exe Investigate if the tray icon API calls can be intercepted

## Block/Allow lists 

The current permissive model is probably not the safest.
`client.Default` and `client.New` should respect the following settings:
Add a config option to 
1. change the nats service port 
> if nats-server is not running on the configured port, try to start it. Otherwise panic
2. specify whether to be permissive or restrictive. 
> Permissive: allow any connection except if specified in block list
> Restrictive: Disallow all connections not specified in allow list

## New yaml parser 

Motivation: windowmanager has workarounds because of this specific issue.

Config files / code would be simpler if there was a generic way to 
parse complex objects from strings. E.g. if a config member implements the 
interface Unyamler with 
`func (u Unyamler) FromString(str string) Unyamler`
`func (u Unyamler) ToString() string`
then e.g. a VKEY should be constructed with the interface calls.
Everything else should be constructed normally.

If we go with a new yaml parser, we can also make it case-insensitive.
As a bonus, we can also do some ultra-unsafe encoding/decoding private fields.
This could also pave the way for some fascinating IPC?

## Bugs 

* Processes created with `driver`'s `System.LaunchProgram` will run as admin if `driver` is running as admin.
- Consider if we need a full-blown CreateProcess implementation which mines the registry and properly controls inherited handles
* `driver` cannot open e.g. a `.png` file after we switched to `ShellExecute`.
- Add handlers to config?
- This is probably better solved with the `CreateProcess` solution?

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
