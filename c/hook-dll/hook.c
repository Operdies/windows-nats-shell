#include <hook.h>
#include <winuser.h>
#define UNUSED(x) (void)(x)
const char *SockAddr = "\\\\.\\pipe\\shellpipe";
// This handle will never be closed.
// It's supposed to run for the duration of the program.
// It looks like the pipe is actually closed when the application exits though.
static HANDLE pipe = NULL;

int WriteToPipe(const char *msg, int len) {
  if (pipe == NULL || pipe == INVALID_HANDLE_VALUE) {
    pipe = CreateFile(SockAddr, GENERIC_WRITE, FILE_SHARE_WRITE, NULL,
                      OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
    if (pipe == INVALID_HANDLE_VALUE) {
      return 0;
    }
  }

  DWORD n = 0;
  WriteFile(pipe, msg, len, &n, NULL);
  return len == n;
}

LRESULT CALLBACK ShellProc(int nCode, WPARAM wParam, LPARAM lParam) {
  if (nCode >= 0){
    char buf[40];
    int len = sprintf(buf, "WH_SHELL,%d,%llu,%lld\n", nCode, wParam, lParam);
    WriteToPipe(buf, len);
  }
  return CallNextHookEx(NULL, nCode, wParam, lParam);
}

