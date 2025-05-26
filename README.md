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

- Pre-existing stream with subjects that can be partitioned
  - In this sample code im using our `iot` stream and partitioning on `avl.broadcast.*`
- Creation of an Elastic Consumer Group using the `cg` tool provided by the library
- Initiatlization of the Consumer Group
- Starting the Consumer
- Adding Member to Consumer Group

### As of now with the current state of the library, above steps can be distributed by the following model of responsibility

#### Tasks for the Administrator of the NATS Cluster

- Pre-existing stream with subjects that can be partitioned
  - In this sample code im using our `iot` stream and partitioning on `avl.broadcast.*`
- Creation of an Elastic Consumer Group using the `cg` tool provided by the library

#### Tasks for the Developer of the consuming application

- Initiatlization of the Consumer Group
- Starting the Consumer
- Adding Member to Consumer Group

---

Further documentation about the inner workings of partitioned consumer groups can be found at the official release repo
https://github.com/synadia-io/orbit.go/tree/main/pcgroups

To use the official library:

```
go get github.com/synadia-io/orbit.go/pcgroups
```
