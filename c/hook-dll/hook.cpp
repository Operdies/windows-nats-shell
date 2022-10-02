#include <hook.h>
#include <winuser.h>
#define UNUSED(x) (void)(x)
const char *SockAddr = "\\\\.\\pipe\\shellpipe";


LRESULT CALLBACK ShellProc(int nCode, WPARAM wParam, LPARAM lParam) {
  UNUSED(wParam);
  UNUSED(lParam);

  if (nCode >= 0){
    char buf[32];
    sprintf(buf, "%d,%llu,%lld", nCode, wParam, lParam);
    WriteToPipeWithRetry(buf, 3);

    // Block opening the start menu when pressing the windows key
    if (nCode == 7) return TRUE;
  }
  return CallNextHookEx(NULL, nCode, wParam, lParam);
}

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
