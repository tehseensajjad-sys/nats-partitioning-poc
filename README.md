### Partitioned Consumer Groups Example

### Usage

- ```
  git clone https://github.com/tehseensajjad-sys/nats-partitioning-poc.git
  ```

- ```
  cd nats-partitioning-poc/
  ```
- Start multiple consumer processes
- ```
  go run main.go
  ```
- In a different terminal session
- ```
  go run main.go
  ```

### General Steps required to consume partitioned messages

1. Pre-existing stream with subjects that can be partitioned

   1a. In this sample code im using our `iot` stream and partitioning on `avl.broadcast.*`

2. Creation of an Elastic Consumer Group using the `cg` tool provided by the library

```
cg elastic create iot poc 'avl.broadcast.*' 1
```

3. Initiatlization of the Consumer Group
4. Starting the Consumer
5. Adding Member to Consumer Group

> A werid but sensible thing about the library is that Step 4 will not begin the flow of messages rather Step 5 is the point where messages will actually start to be consumed and sent to the message handler function

As of now with the current state of the library, above steps can be distributed by the following model of responsibility

| Role                    | Steps  |
| ----------------------- | ------ |
| Administrator           | 1 & 2  |
| Consuming app Developer | 3 to 5 |

---

### Internals

- A replica WorkQueue stream is created which sources messages from the target stream.

- Each consuming member is a seperate ephemeral **NATS Consumer**

- Members are tracked using NATS Key Value Store (`nats kv ls`)
- Two members with the same name form a Active Passive relation where one will consume messages and the other will wait indefinitely, upon the Active member to crash the passive one will take over and start consuming

---

Further documentation about the inner workings of partitioned consumer groups can be found at the official release repo
https://github.com/synadia-io/orbit.go/tree/main/pcgroups

To use the official library:

```
go get github.com/synadia-io/orbit.go/pcgroups
```
