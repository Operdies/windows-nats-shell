# diary

## Create rofi integration
  Investigate what is possible for rofi modules
  rofi -dmenu only supports text. This can be used to support many cases,
  but makes scripting difficult if the output should match partial strings, e.g. the hwnd from a 'SetFocus' request
  With real integration, the launcher can also ship real icons
## GetProcessList nats endpoint
  GetProcessList should contain extenteded, useful information 
  about the process. GetWindows is probably still different enough
  that they can both coexist.
  Evaluate if NewProcess events are necessary
## NATS process IO 
  IO should just be done using nats. So the default stdout / stderr for a service will just be a byte buffer which publishes logs 
  with the service name as a subject. Then services can be NATS agnostic
  stdin can also be nats for that matter. 
## NATS log service
  logs for services could be stored in a sqlite database
  and be queried using NATS. Then there would be no need 
  for a console window to host the shell
## Kill menu
  Make an API to kill a process by its handle
## Shortcut manager
  Implement keyboard hook in GO in order to intercept keys that are mapped.
  Add a configuration override to pass through mapped keys
## Background 
  API for posting data (images or text) to the background. TBD: pre-configured zones or defined per request / client?
## Window manager 
  Some (toggle-able) way to automatically tile windows. BSPC inspired layouts
  Cycle between avant-garde fanning strategy, tall mode, and monocle mode
  Disable borders on all windows (maybe with hotkey to enable currently focused window?)
  Mouse controls to resize / reposition windows (win+left/right drag)
## Steam integration 
  Launch / install games
  Investigate what integrations exist / are possible
## Rofi but for windows
  Would make the shell usable without a linux driver. The shortcut manager needs to support input/output. Then the rofi implementation can respond using nats 

## Thought cabinet
  > Service namespaces?
  Services where multiple instances make sense should really use namespaces. The clients should probably support namespaces in some way 
  The environment variable containing the service name could just be prepended as a subject where it makes sense. But then 
  then e.g. Requester clients also have to require the namespace as an input. Do clients care which instance responds?

  > How can input/output be implemented? Named actions?
  `... payload: { hwnd: $action1.hwnd, command: $action2.command }` ?
  Then actions with no dependencies can simply be `publish`ed. A dependency tree must be built of other actions.
  There should probably be a panic during startup if circular dependencies are detected.
  There should also be a startup panic if collisions are detected.

  > Should all APIs require complex type inputs ?
  Should all e.g. Windows APIs require a window as input? If an action only requires e.g. `hwnd` to be set, the input can just be `hwnd: 1234`
  More generally, should all X apis take an input of the same form as their outputs? It would make it possible to return meaningful errors.
  Rofi integration is a prerequisite for this, because the alternative is way too complicated on the scripting side.


  > When are stdout/stderr logs consumed?
  It does not seem like a viable solution to store all logs in memory indefinitely.
  It would make the most sense if logs are published immediately as they arrive, 
  and then discarded by the shell. If anyone cares about the logs they must subscribe to them.
  If the service name is `$subject.stdout` then a service can easily get all logs with `*.stdout`.
  There should be a database service which collects the logs.
