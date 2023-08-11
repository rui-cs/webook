# webook

项目目录层级结构：

+ service : 代表领域服务(domain service)。组合各种repository和domain，偶尔组合别的service，共同完成一个业务功能。
+ repository : 代表领域对象的存储。只代表数据存储，不代表数据库
  + dao : 代表数据库操作
+ domain : 代表领域对象

```mermaid
sequenceDiagram
participant A as web(和HTTP打交道)
participant B as service(主要业务逻辑)
participant C as repository(数据存储抽象)
participant D as dao(数据库操作)

A->>B: data
activate A
activate B
B->>C : data
activate C
C->>D : data
activate D
D->>C : data
deactivate D
C->>B : data
deactivate C
B->>A: data
deactivate B
deactivate A
```

domain.User 是业务概念

dao.User 直接映射数据库中的表