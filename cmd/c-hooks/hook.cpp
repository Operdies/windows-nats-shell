#include <hook.hpp>
#include <nats.h>

int publish() {
  printf("Publishes a message on subject 'foo'\n");
  fflush(stdout);
  natsConnection *conn = NULL;
  natsMsg *msg = NULL;
  natsStatus s;

  // Creates a connection to the default NATS URL
  s = natsConnection_ConnectTo(&conn, NATS_DEFAULT_URL);
  if (s == NATS_OK) {
    const char data[] = {104, 101, 108, 108, 111, 33};

    // Creates a message for subject "foo", no reply, and
    // with the given payload.
    s = natsMsg_Create(&msg, "hello", NULL, data, sizeof(data));
  }
  if (s == NATS_OK) {
    // Publishes the message on subject "foo".
    s = natsConnection_PublishMsg(conn, msg);
  }

  // Anything that is created need to be destroyed
  natsMsg_Destroy(msg);
  natsConnection_Destroy(conn);

  // If there was an error, print a stack trace and exit
  if (s != NATS_OK) {
    nats_PrintLastErrorStack(stderr);
    exit(2);
  }

  return 0;
}

// LRESULT CALLBACK ShellProc(int nCode, WPARAM wParam, LPARAM lParam) {
//   return 0;
// }
