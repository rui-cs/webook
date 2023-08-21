# webook

## **项目目录层级结构**

+ web : web 中的handler负责和HTTP有关的内容
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



**gin**

web 框架采用 https://gin-gonic.com/zh-cn/

gin middleware 库： https://github.com/gin-gonic/contrib 使用`Engine.Use`。



**middleware**

middleware 是 Go 这里用得比较多的说法，在别的语言里面可能叫做 plugin、handler、filter、 interceptor。

```mermaid
graph LR
	subgraph middleware
		D[middleware1]
		E[middleware2]
		F[middleware...]
	end
	
 request-->D-->E-->F-->业务逻辑
```

请求都要经过这些 middleware，所以适合用来解决一些所有业务都关心的东西。比如说里的跨域问题，注册的所有的路由都需要解决。也叫做 AOP(Aspect-Oriented Programming) 解决方案。



## **密码加密**

可以选择在不同的层加密

+ service 加密：加密是一个业务概念
+ repository 加密：加密是一个存储概念
+ dao 加密：加密是一个数据库概念
+ domain 加密 ：加密是一个业务概念，但应该是“用户（User）”自己才知道怎么加密

选择servcie加密。

常见的加密算法

+ md5之类的哈希算法
+ 在第一点基础之上引入盐(salt)，或者多次哈希等
+ PBKDF2、BCrypt这一类随机盐值的加密算法，同样的文本加密后的结果都不一样

选择BCrypt加密，号称最安全的加密算法。优点有：

+ 不需要自己生成盐值
+ 不需要额外存储盐值
+ 可以通过控制cost来控制加密性能
+ 同样的文本，加密后的结果不同

bcrypt加密之后无法破解，只能同时比较加密之后的值来确定两者是否相等。



## **实现登录功能**

登录本身分为两部分

+ 实现登录功能 (/users/login接口)
+ 登录态校验 (Cookie, Session)



浏览器会把 Cookie (是一些数据，格式是键值对) 存储到本地，这样不太安全。

Cookie 使用字段

+ 响应头字段 Set-Cookie 
+ 请求头字段 Cookie

Cookie的关键配置

+ Max-Age和Expires : 过期时间。Max-Age 单位是秒，浏览器优先采用Max-Age计算失效期。
+ Domain和Path : Cookie 可以用在什么域名和路径下。设定原则：最小化原则
  + “Domain”和“Path”指定了 Cookie 所属的域名和路径，浏览器在发送 Cookie 前会从 URI 中提取出 host 和 path 部分，对比Cookie 的属性。如果不满足条件，就不会在请求头里发送 Cookie。
+ HttpOnly : 设置为true时，浏览器上的JS代码将无法使用这个Cookie。防止“跨站脚本”（XSS）攻击窃取数据，提升Cookie安全性。
+ SameSite : 是否允许跨站发送Cookie。防范“跨站请求伪造”（XSRF）攻击，提升Cookie安全性。
+ Secure : 只能用于HTTPS协议。提升Cookie安全性。

Cookie应用：

+ Cookie 最基本的一个用途就是身份识别，保存用户的登录信息，实现会话事务。
+ Cookie 的另一个常见用途是广告跟踪。

Cookie总大小不能超过4K。

Cookie名称来源Magic Cookie，含义不透明的数据。

注意，Cookie 并不属于 HTTP 标准。



因为Cookie不安全，所以关键数据可以存储到Session中，并保存在后端。访问系统的时候带上session id，后端根据session id识别访问者身份。session id可放在：

+ Cookie
+ Header
+ 查询参数，即 ?sid=XXX

session中的数据存储在 `store` 结构中。（https://github.com/gin-contrib/sessions）



store选择：

+ 单机单实例部署，可选择memstore，基于内存方式实现。
+ 多实例部署，可选择redis。



redis实现store中需要

+ authentication：身份认证
+ encryption：数据加密

信息安全的三个核心概念：authentication，encryption，authorization（授权，即权限控制）



gin-session中间件的各种类型实现是面向接口编程的。可以自由切换。当你在设计核心系统的时候，或者你打算提供什么功能给用户的时候，一定要问问自己，将来有没有可能需要不同的实现。



session 刷新

需要在用户持续使用网站时，刷新过期时间。

刷新策略

+ 每次访问都刷新：性能差，对redis影响大
+ 快过期的时候刷新：快过期的时候用户没访问无法刷新
+ 固定间隔时间刷新：比如每分钟内第一次访问都刷新
+ 使用长短token

设置session有效期为60s，在登录校验的middleware中，登录校验之后顺手刷新。刷新规则是，如果没有设置过session的update_time 或者当前时间超过update_time 10s，则重置session有效期。



登录状态保持多久比较好？ 

登录状态保持多久比较好？也就是，一次登录之后，要隔多久才需要继续登录？ 

答案是取决于你的产品经理，也取决于你系统其它方面的安全措施。 

简单来说，就是如果你有别的验证用户身份的机制，那么你就可以让用户长时间不需要登录。

上述60s和10s都可根据实际情况修改。



## **/users/profile 和 /users/edit 接口设计**

/users/profile 接口设计

返回信息：邮箱，用户名，生日，个人简介

/users/edit 接口设计

可修改用户名、生日、个人简介

需校验生日格式，用户名唯一

返回错误：系统错误 / 用户名重复 / 生日格式错误返回 http code 400 Bad Request

实现：旧版设计是要从email定位用户，后修改为从session中拿出userID定位用户。按照web-->service-->repository-->dao层次操作数据库即可。

测试结果：截图在test文件



## JWT

除了使用 gin-session middleware 保持和校验登录态，也可以用JWT(JSON Web Token)。

JWT主要用于身份认证，即登录。

基本原理：通过加密生成一个 token，而后客户端每次访问的时候都带上这个 token。



JWT 简介 

它由三部分组成： 

+ Header：头部，JWT 的元数据，也就是描述这个token 本身的数据，一个 JSON 对象。 
+ Payload：负载，数据内容，一个 JSON 对象。 
+ Signature：签名，根据 header 和 token 生成。



如何进行接入改造？

使用 JWT 原始 API ：go get github.com/golang-jwt/jwt/v5 

在登录过程中，使用 JWT 也是两步： 

+ JWT 加密和解密数据
+ 登录校验

过程：

+ 在 Login 接口中，登录成功后生成 JWT token。 
  + 在 JWT token 中写入数据。 
  + 把 JWT token 通过 HTTP Response Header `x-jwt-token` 返回。 

+ 改造跨域中间件，允许前端访问 `x-jwt-token` 这个响应头。 

+ 接入 JWT 登录校验的 Gin middleware。 
  + 读取 JWT token。 
  + 验证 JWT token 是否合法。 

+ 下发HTTP请求时要携带 JWT token。
+ 从session中获取userID的地方需要改为从JWT中获取userID，如 /users/profile，/users/edit 接口

```mermaid
sequenceDiagram
participant A as 前端
participant B as JWT登录校验
participant C as Login接口
participant D as 其他接口

A->>C: 登录
activate A
activate C
C->>C : 生成token

C->>A : x-jwt-token:xx
deactivate C
deactivate A

A->>B: x-jwt-token:xx
activate A
activate B


B->>B : token没问题
deactivate B
activate D
B->>D:  
D->>D: 
D->>A: 
deactivate D
deactivate A
```





JWT 的优缺点 

和 Session 比起来，优点： 

+ 不依赖于第三方存储。 
+ 适合在分布式环境下使用。 
+ 提高性能（因为没有 Redis 访问之类的）。 

缺点： 

+ 对加密依赖非常大，比 Session 容易泄密。 
+ 最好不要在 JWT 里面放置敏感信息。



混用 JWT 和 Session 机制 

前面 JWT 限制了我们不能使用敏感数据，那么你真有类似需求的时候，就可以考虑将数据放在 “Session”里面。 

基本的思路就是：你在 JWT 里面存储你的 userID，然后用 userID 来组成 key，比如说 user.info:123 这种 key，然后用这个 key 去 Redis 里面取数据，也可以考虑使用本地缓存数据。



## 保护系统

保护系统要考虑两方面

+ 正常用户会不会搞崩你的系统？
+ 如果有人攻击你的系统，能否撑得住？

现在系统最明显的漏洞

+ 任何人都能注册
+ 任何人都能登录

限流是常见的保护系统的手段。限制每个用户每秒最多发送固定请求的数量。

问题

+ 怎么判定请求是某个用户的？未登录成功时，不知道用户是谁
+ 怎么确定限流阈值？

第一个问题：限流对象，可以用IP / MAC地址 / 设备标识符(CPU序列号) 。

第二个问题：阈值问题，理论上来说，这应该是通过压测来得到的。比如说你压测整个系统，发现最多只能撑住每秒 1000 个请求，那么阈值就是 1000。而我们是针对个人，搞不了压测。所以可以凭借经验来设置，比如说我们正常人手速，一秒钟撑死一个请求，那么就算我们考虑到共享 IP 之类的问题，给个每秒 100 也已经足够了。

实现：用redis做限流



## 增强登录安全

这种实现方式有个问题，不管是用 JWT 还是 Session，一旦被攻击者拿到关键的 JWT 或者 ssid，攻击者就能假冒你。 HTTPS 可以有效阻止攻击者拿到你的 JWT 或者 ssid。 但是如果你电脑中了病毒，那 HTTPS 就无能为力。

在用户登录校验过程中，得进一步判断，用这个 JWT/ssid 的人是不是原本登录的那 个人。目前做得好的都是使用二次验证，也就是发邮件、 发短信等。但是也有一些比较初步但也好用的手段，那就是用登录的辅助信息来判断。



登录的时候，记录当时登录的一些额外信息。比如说:

+ 使用的浏览器：对应到HTTP的User-Agent头部
+ 硬件信息：手机APP比较多见。

在登录校验的时候，比较一下你当次请求的这些辅助信 息和上一次的信息，不一样就认为有风险

问题:能不能用 IP? 不能，IP随时可能会切换



需要改造两个地方：

+ Login接口，在JWTtoken里面带上User-Agent信息。
+ JWT登录校验中间件，在里面比较User-Agent。



## 跨域问题处理

什么是跨域请求？

协议、域名和端口任意一个不同，都是跨域请求

从postman发送本地请求不会遇到跨域问题，但是从浏览器发送请求就可能会遇到。浏览器有这样一个机制，会自动发送preflight请求。



CORS middleware (https://github.com/gin-gonic/contrib/tree/master/cors)

+ AllowOriginFunc : 哪些来源是允许的。
+ AllowHeader : 业务请求中可以带上的头。 
+ AllowCrendentials : 是否允许带上用户认证信息(比如 cookie)。
+ ExposedHeaders : 允许显示的响应头（这样前端才能拿到）



/users/login 接口测试跨域问题解决效果

request header 中有

```
Host: localhost:8080
Origin: http://localhost:3000
Referer: http://localhost:3000/
```

response header 中有

```
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: http://localhost:3000
Access-Control-Expose-Headers: X-Jwt-Token
Vary: Origin
```



跨域问题要点

+ 跨域问题是因为发请求的`协议+域名+端口`和接收请求的`协议+域名+端口`对不上。比如说从 localhost:3000 发到 localhost:8080 上。
+ 解决跨域问题的关键是在 preflight 请求里面告诉浏览器自己愿意接收请求。
+ Gin 提供了解决跨域问题的 middleware，可以直接使用。
+ middleware 是一种机制，可以用来解决一些所有业务都关心的问题，使用 Use 方法来注册middleware



## Kubernetes 入门

[示例](https://github.com/rui-cs/go-learning/tree/main/kubernetes)



初学记住几个基本概念

+ Pod ：实例
+ Service ： 逻辑上的服务
+ Deployment ： 管理Pod

Pod 和 Service 最简单的理解方式：假如说你有一个 Web 应用，部署了三个实例，那么就是一个 Web Service，对应了三个 Pod。

Deployment 最好的理解方式：你跟运维说要保证我的 Web 有三个实例，少了运维就重启一个，多了运维就删除一个，运维就是那个 Deployment。



```mermaid
graph LR
	subgraph pod
		D[pod0]
		E[pod1]
		F[pod2]
	end
	
 service-->pod
 Deployment-->pod
```

**配置说明**

k8s文档：https://kubernetes.io/zh-cn/docs/home/

K8s 简单理解就是一个配置驱动的，或者元数据驱动，或者声明式的框架

编写Deployment

+ apiVersion
+ spec，可以理解为说明书
  + replicas
  + selector : 筛选器，在所有的pod中，要管理哪三个pod。可以用matchLabels (根据给出的label值筛选) 和matchExpressions (根据表达式筛选)。
  + template : 该怎么创建每个pod。Kind不同时template也不同。

编写Service

只有 Deployment 无法从外面访问，需要将 Pod 封装为一个逻辑上的服务，即 Service。

+ apiVersion
+ kind
+ metadata
+ spec
  + type：可选择负载均衡
  + ports：端口

goland有k8s的插件，编写配置文件很方便



**安装**

直接在docker desktop中使用k8s，另外需要安装一个kubectl （设备：macbook）

docker desktop中如果启动k8s一直是starting状态，可能是网络问题无法拉取镜像

记得切换Kubernetes运行上下文至 docker-for-desktop

```shell
# 切换
kubectl config use-context docker-desktop

# 查看
kubectl get node
NAME             STATUS   ROLES           AGE   VERSION
docker-desktop   Ready    control-plane   70m   v1.27.2

# 查看
kubectl describe node docker-desktop
```



**用 Kubernetes 部署 web 服务器**

入门例子：部署三个web服务器实例，需要一个service，一个deployment，三个pod，每个pod是一个实例

```mermaid
graph LR
	subgraph pod
		D[pod0 : webook]
		E[pod1 : webook]
		F[pod2 : webook]
	end
	
 /hello-->webook-service-->pod
```



**用 Kubernetes 部署 Redis**

仅部署单机版redis，不考虑持久化问题。

port、nodePort 和 targetPort 的含义

+ port : 是指 Service 本身的，比如在 Redis 里面连接信息用的就是 demo-redis-service:6379
+ nodePort : 是指在 K8s 集群之外访问的端口，比如说执行 redis-cli -p 30379
+ targetPort : 是指 Pod 上暴露的端口

<img src="./pic/k8s三种Port.jpg" alt="k8s三种Port" style="zoom:30%;" />





**用 Kubernetes 部署 MySQL**

部署MySQL与前面部署web服务器和redis不同的一点是，其需要数据持久化。

在 K8s 里面，存储空间被抽象为 PersistentVolume(持久化卷)。

如何理解 PersistentVolume？

+ 作为 K8s 的设计者，不知道容器里面运行的会是什么东西，需要怎么存储，管不了。
+ 从现实中来看，有各种设备用于存储数据，比如说机械硬盘、SSD，又比如说各种封装、各种文件协议，也管不了。 

最终你只能考虑提供一个抽象，让具体的实现去管了。



在已有service yaml和deployment yaml基础之上，分三步走：

+ Deployment yaml文件中加template

  在 template 里面，关键是 `spec.containers.volumeMounts` 和 `volumes`。

  +  `spec.containers.volumeMounts` 含义是挂载到容器的哪个地方
  +  `volumes` ：含义是这里挂载的东西究竟是什么

  ```yaml
  spec:
    replicas: 1
    selector:
      matchLabels:
        app: webook-mysql
    template:
      metadata:
        name: webook-mysql
        labels:
          app: webook-mysql
      spec:
        containers:
          - name: webook-mysql
            image: mysql:8.0
            imagePullPolicy: IfNotPresent
            env:
              - name: MYSQL_ROOT_PASSWORD
                value: root
            volumeMounts:
                # 与 mysql 的数据存储位置对应
              - mountPath: /var/lib/mysql
                # 确定具体用pod中的哪个 volume
                name: mysql-storage
            ports:
              - containerPort: 3306
        restartPolicy: Always
        #  POD 中有哪些volume
        volumes:
          - name: mysql-storage
            persistentVolumeClaim:
              claimName: webook-mysql-claim
  ```

  上面的配置中，含义是在 MySQL 里面挂载一 个目录 `/var/lib/mysql`。当容器读写这个目录的时候，实际上读写的是 `mysql-storage`。

  而 `mysql-storage` 究竟是什么，被一个叫做 `webook-mysql-claim` 的东西声明了。

+ 加PersistentVolumeClaim yaml文件

  一个容器需要什么存储资源，是通过 PersistentVolumeClaim 来声明的。

  比如说，我现在是 MySQL，我就需要告诉 K8s 我需要一些什么资源。K8s 就会为我找到对应的资源。

+ 加PersistentVolume yaml文件

  持久化卷，表达我是一个什么样的存储结构。

  所以，PersistentVolume 是存储本身说我有什么特性，而 PersistentVolumeClaim 是用的人告诉 K8s 说他需要什么特性。

  + storageClass : PersistentVolumeClaim 和 PersistentVolume 的yaml文件中的 storageClassName 要能对上

  + accessMode：访问模式。设想，如果你设计一个存储的东西，你是不是要考虑，我这个东西是只读还是只写？是允许 一个人访问，还是允许很多人访问？这就是由 accessMode 来控制的。

    在 PersistentVolume 里面，accessMode 是说明这个 PV 支持什么访问模式。

    在 PersistentVolumeClaim 里面，accessMode 是说明这个 PVC 需要怎么访问。

    accessMode有以下选项：

    + ReadWriteOnce : 只能被挂在到一个 Pod，被它读写。
    + ReadOnlyMany : 可以被多个 Pod 挂载，但是只能读。
    + ReadWriteMany : 可以被多个 Pod 挂载，它们都能读写。



![pv和pvc](./pic/pv和pvc.jpg)



**用 Kubernetes 部署 Nginx**

什么是Ingress？

用一句话来说，Ingress 代表路由规则。前端发过来的各种请求，在经过 Ingress 之后会转发到特定 的 Service 上。和 Service 中的 LoadBalancer 比起来，Service 强调的是将流量转发到 Pod 上，而 Ingress 强调的是发送到不同的 Service 上。

<img src="./pic/什么是ingress.jpg" alt="什么是ingress" style="zoom:25%;" />



Ingress 和 Ingress controller

一个 Ingress controller 可以控制住整个集群内部的所有 Ingress(符合条件的 Ingress)。

或者这么说:

+ Ingress 是你的配置，配置了一些路由规则
+ Ingress controller 是执行这些配置的，实际执行转发的

站在 K8s 设计者的角度，K8s 只需要一份路由规则说明(Ingress)，至于谁来执行这个路由规则，怎么执行这个路由规则，不关心。

<img src="./pic/Ingress 和 Ingress controller.jpg" alt="Ingress 和 Ingress controller" style="zoom:20%;" />



**安装 helm 和 ingress-nignx**

安装 helm

https://helm.sh/docs/intro/install/

```shell
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh
```

也可以直接下载安装包

使用 helm 安装 ingress-nginx，运行

```shell
helm upgrade --install ingress-nginx ingress-nginx --repo https://kubernetes.github.io/ingress-nginx --namespace ingress-nginx --create-namespace
```

安装完 ingress-nginx 相当于把nginx安装好了， ingress-nginx相当于ingress controller。



编写 Ingress 配置文件

关键点

+ apiVersion 是一个新的东西，networking.k8s.io/v1，文档里面会告诉你这个值应该是什么。 
+ spec.ingressClassName 需要指定为 nginx，如果用别的 Ingress，也要指定对应的名字。
+ rules : 配置的你的转发规则，基本上就和你平时配置nginx 差不多。








## **参考**

+ 主要内容来自大明训练营
+ 图解HTTP 6.7 小节
+ 透视HTTP协议 19课