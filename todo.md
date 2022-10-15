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
## NATS log service
  logs for services could be stored in a sqlite database
  and be queried using NATS. Then there would be no need 
  for a console window to host the shell
## Kill menu
  Make an API to kill a process by its handle
## Shortcut manager
  Implement keyboard hook in GO in order to intercept keys that are mapped.
  Add a configuration override to pass through mapped keys
  Fix bug where the same keyevents are duplicated in Chrome 
  -- In general I think this happens when key events are posted to multiple receivers.
  I'm not sure what the best way to tackle this is. Do all keyboard listeners implement 
  throttling, or does the event publisher throttle keys?
  -- Should there be a different subject for "give me real keystrokes" and "give me all key events"?
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
## Custom configs 
  Currently, the workaround for a service to have a custom config is the custom key, and a helper method to remarshal the config.
  We can avoid the `custom` key if the `Config` endpoint returns the entire config (and not just the part the shell understands)
  Then the requesting service should say `client.GetConfig[MyConfig](requester)`, and define `MyConfig` like 
  ```go 
  type MyConfig struct {
		  base shell.Service 
		  MySettings string
	  }
  ```
  This would be much simpler
## C callbacks
  The C callbacks are identical except for a prefix.
  It would make sense if Microsoft made the nCode value unique between the different event types. 
  It would also make sense if they designed them such that e.g. the range `x < nCode < y` codes all belong to event `z`
  Investigate if these assumptions are accuracte and rewrite callbacks.

  Bonus: Check if the callback can be done in GO to skip the named pipe business.
  I read a post about some GO C exports that sort of looked like it would be possible. Then the named pipes can 
  be scrapped.

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

> Be aware that the WH_MOUSE, WH_KEYBOARD, WH_JOURNAL*, WH_SHELL, and low-level hooks can be called on the thread that installed the hook rather than the thread processing the hook. For these hooks, it is possible that both the 32-bit and 64-bit hooks will be called if a 32-bit hook is ahead of a 64-bit hook in the hook chain. For more information, see Using Hooks.
> https://github.com/moutend/go-hook/blob/develop/pkg/keyboard/keyboard_windows.go 
> please immediately get rid of the hooks running in foreign threads??
> please also investigate if WH_KEYBOARD_LL can do anything WH_KEYBOARD can't
