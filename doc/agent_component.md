```uml
@startuml
actor admin

node "controller" {
  frame "central agent" {
    admin --> ()cli
    ()cli -> [local_agent_manager]
    [monitoring_manager]
  }
  () OpenStack_API
}

[local_agent_manager] ---> ()broker : pub

node "compute01"{
  frame "local agent" {
    [API_kicker]
    [info_adder]
    [process_controller]
    [file_manager]
  }
  [local_agent_manager] --> [file_manager] : send setting(gRPC/zeroMQ/etc)
  [file_manager] -> ()broker : sub
  [file_manager] --> [collector]
  [process_controller] --> [collector] : stop & start
  [file_manager] --> [notificator]
  [file_manager] --> [analytics engine]
  database "in-memory DB"
  [API_kicker] --> () OpenStack_API
  [API_kicker] --> [in-memory DB] : insert inventory
  [in-memory DB] -> [info_adder]
  [info_adder] --> [notificator] : add host/vm info
}

@enduml
```
