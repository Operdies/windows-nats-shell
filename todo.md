# Some tasks idk

There are a lot of APIs already that can be put to good use

# Create rofi integration
  Investigate what is possible for rofi modules
  rofi -dmenu only supports text. This can be used to support many cases,
  but makes scripting difficult if the output should match partial strings, e.g. the hwnd from a 'SetFocus' request
  With real integration, the launcher can also ship real icons
# GetProcessList nats endpoint
  GetProcessList should contain extenteded, useful information 
  about the process. GetWindows is probably still different enough
  that they can both coexist.
  Evaluate if NewProcess events are necessary
# Kill menu
  Make an API to kill a process by its handle
# Shortcut manager
  Something similar to sxhkd?
  How to avoid posting launch keys to application?
  Does this app actually need to install its own hook?
  Could such a hook run in GO since keyboard hooks don't have the 'other module' requirement?
  Can we drop the NATS keylogger?
# Background 
  Investigate freezes ?
  Configurable shaders 
  API for posting data (images or text) to the background. 
  TBD: pre-configured zones or defined per request / client?
# Window manager 
  Some (toggle-able) way to automatically tile windows. BSPC inspired layouts
  Cycle between avant-garde fanning strategy, tall mode, and monocle mode
  Disable borders on all windows (maybe with hotkey to enable currently focused window?)
  Mouse controls to resize / reposition windows (win+left/right drag)
# Steam integration 
  Launch / install games
  Investigate what integrations exist / are possible
# Rofi but for windows
  Would make the shell usable without a linux driver
