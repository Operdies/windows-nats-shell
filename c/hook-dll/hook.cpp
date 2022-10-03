#include <hook.h>
#include <winuser.h>
#define UNUSED(x) (void)(x)
const char *SockAddr = "\\\\.\\pipe\\shellpipe";

int WriteToPipe(const char *msg) {
  HANDLE pipe = NULL;

  pipe = CreateFile(SockAddr, GENERIC_WRITE, FILE_SHARE_WRITE, NULL,
                    OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
  if (pipe == INVALID_HANDLE_VALUE) {
    return 0;
  }

  DWORD n = 0;
  DWORD len = strlen(msg);
  WriteFile(pipe, msg, len, &n, NULL);
  CloseHandle(pipe);
  return len == n;
}


int WriteToPipeWithRetry(const char *msg, int lim) {
  for (int i = 0; i < lim; i++) {
    if (WriteToPipe(msg)) {
      return 1;
    }
    Sleep(1);
  }
  return 0;
}

LRESULT CALLBACK KeyboardProc(int nCode, WPARAM wParam, LPARAM lParam){
  if (nCode >= 0){
    char buf[40];
    sprintf(buf, "WH_KEYBOARD,%d,%llu,%lld", nCode, wParam, lParam);
    WriteToPipeWithRetry(buf, 3);
  }
  return CallNextHookEx(NULL, nCode, wParam, lParam);
}

LRESULT CALLBACK CBTProc(int nCode, WPARAM wParam, LPARAM lParam) {
  if (nCode >= 0){
    char buf[40];
    sprintf(buf, "WH_CBT,%d,%llu,%lld", nCode, wParam, lParam);
    WriteToPipeWithRetry(buf, 3);
  }
  return CallNextHookEx(NULL, nCode, wParam, lParam);
}

LRESULT CALLBACK ShellProc(int nCode, WPARAM wParam, LPARAM lParam) {
  if (nCode >= 0){
    char buf[40];
    sprintf(buf, "WH_SHELL,%d,%llu,%lld", nCode, wParam, lParam);
    WriteToPipeWithRetry(buf, 3);
  }
  return CallNextHookEx(NULL, nCode, wParam, lParam);
}

