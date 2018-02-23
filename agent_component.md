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

node "compute01"{
  frame "local agent" {
    [file_manager]
    [process_controller]
    [API_kicker]
    [info_adder]
  }
  [policy manager] --> [file_manager] : send setting files
  [file_manager] --> [collector]
  [process_controller] --> [collector] : stop & start
  [file_manager] --> [notificator]
  [file_manager] --> [analytics engine]
  database "in-memory DB01"
  [API_kicker] --> () OpenStack_API
  [info_adder] --> [notificator] : add host/vm info
}

@enduml
```
