#include <stdio.h>
#include <windows.h>
#include <winnt.h>


extern "C" {
// __declspec(dllexport) int WriteToPipeWithRetry(const char* msg, int lim);
__declspec(dllexport) LRESULT CALLBACK ShellProc(int nCode, WPARAM wParam, LPARAM lParam);
__declspec(dllexport) LRESULT CALLBACK CBTProc(int nCode, WPARAM wParam, LPARAM lParam);
__declspec(dllexport) LRESULT CALLBACK KeyboardProc(int nCode, WPARAM wParam, LPARAM lParam);
}
