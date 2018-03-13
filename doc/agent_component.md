```uml
@startuml
actor admin

node "controller" {
  frame "central agent" {
    admin --> ()cli
    ()cli -> [policy manager]
    [local_agent_manager]
  }
  () OpenStack_API
}

[policy manager] ---> ()broker : pub

node "compute01"{
  frame "local agent" {
    [file_manager]
    [process_controller]
    [API_kicker]
    [info_adder]
  }
  [policy manager] --> [file_manager] : send setting(gRPC/zeroMQ/etc)
  [file_manager] -> ()broker : sub
  [file_manager] --> [collector]
  [process_controller] --> [collector] : stop & start
  [file_manager] --> [notificator]
  [file_manager] --> [analytics engine]
  database "in-memory DB"
  [API_kicker] --> () OpenStack_API
  [API_kicker] --> [in-memory DB] : insert inventory
  [info_adder] --> [notificator] : add host/vm info
}

@enduml
```
