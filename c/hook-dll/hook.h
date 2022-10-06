#include <stdio.h>
#include <windows.h>
#include <winnt.h>


LRESULT CALLBACK ShellProc(int nCode, WPARAM wParam, LPARAM lParam);
LRESULT CALLBACK CBTProc(int nCode, WPARAM wParam, LPARAM lParam);
LRESULT CALLBACK KeyboardProc(int nCode, WPARAM wParam, LPARAM lParam);
