# Networking Notes

This document contains networking concepts relevant to KuFlow.

---

# What Is a Proxy?

A proxy server receives requests from clients and forwards them to another server.

```text
Client -> Proxy -> Target Server
```

KuFlow is a **forward proxy**.

---

# Forward Proxy vs Reverse Proxy

<Table columnSizing="equal" rowDivider={1}><Table.Row header><Table.Cell>Forward Proxy</Table.Cell><Table.Cell>Reverse Proxy</Table.Cell></Table.Row><Table.Row><Table.Cell>Represents the client</Table.Cell><Table.Cell>Represents the server</Table.Cell></Table.Row><Table.Row><Table.Cell>Used for outbound traffic</Table.Cell><Table.Cell>Used for inbound traffic</Table.Cell></Table.Row><Table.Row><Table.Cell>Example: browser proxy</Table.Cell><Table.Cell>Example: Nginx in front of a website</Table.Cell></Table.Row></Table>

KuFlow implements a **forward proxy**.

---

# HTTP Request Structure

Example:

```http
GET /users/octocat HTTP/1.1
Host: api.github.com
User-Agent: curl/8.0
```

Important parts:

* Method
* Path
* Headers
* Body

KuFlow will mainly collect metadata about these requests.

---

# HTTP Response Structure

Example:

```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 123

{ ... }
```

Important parts:

* Status code
* Headers
* Body size

---

# TCP Connection

HTTP is usually transported over TCP.

Simplified flow:

```text
Client
   |
   | SYN
   v
Server
   |
   | SYN-ACK
   v
Client
   |
   | ACK
   v
Connection Established
```

The proxy sits between the client and the upstream server.

---

# DNS Resolution

When a client requests:

```text
https://github.com
```

The hostname must be resolved to an IP address.

KuFlow may later collect DNS timing metrics.

---

# HTTPS

HTTPS = HTTP over TLS.

The first version of KuFlow will support plain HTTP.

HTTPS support will be added later.

---

# Keep-Alive

Without Keep-Alive:

```text
Request -> Open TCP -> Close TCP
```

With Keep-Alive:

```text
Open TCP
  -> Request 1
  -> Request 2
  -> Request 3
Close TCP
```

This greatly improves performance.

---

# Latency

For KuFlow:

```text
latency = time(response received) - time(request received)
```

The metric will be stored as:

```text
latency_ms
```

---

# Future Topics

To be documented later:

* CONNECT method
* TLS handshake
* HTTP/2
* HTTP/3
* WebSockets
* Connection pooling
* Load balancing
