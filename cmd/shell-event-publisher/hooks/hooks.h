#include <Windows.h>

static const char* hookDll = "C:\\Users\\alexw\\repos\\minimalist-shell\\bin\\libhook.dll";
static HMODULE hModule;
static HOOKPROC hookProc;
static HHOOK hhk;


static void RegisterHook(){
  hModule = LoadLibrary(hookDll);
  hookProc = (HOOKPROC)GetProcAddress(hModule, "ShellProc");
  hhk = SetWindowsHookExA(WH_SHELL, hookProc, hModule, 0);
}

static void UnregisterHook(){
  UnhookWindowsHookEx(hhk);
}

