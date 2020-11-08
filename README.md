## 简介
本项目是小米云服务网盘功能的命令行客户端，主要实现了登录、列表、文件下载和上传、相册下载、创建目录、删除、分享等常用功能。项目完全采用Go语言开发，但是目前主要在Linux环境下测试使用，Windows和Mac暂未测试，理论上应该是没问题的，感兴趣的同学可以试一下，如果有问题可以提issue。
```
Go@MiCloud:$ h
NAME:
   micloud - MiCloud Third Party Console Client Written In Golang

USAGE:
   micloud command [command options] [arguments...]

COMMANDS:
   login          登录小米云服务账号
   ls             列表当前目录所有文件和文件夹
   download       下载文件或者文件夹
   cd             改变当前目录，例如：cd movies
   upload         上传文件或者文件夹
   share          获取一个公共分享链接
   rm             删除文件或者文件夹，即放入回收站
   mkdir          创建目录
   tree           打印树型目录结构
   jobs           展示后台当前所有下载和上传任务
   quit, exit     退出应用
   lsAlbum        列出所有相册
   downloadAlbum  下载相册照片文件
   help, h        Shows a list of commands or help for one command

OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## 安装
```shell
wget -c https://github.com/wangbjun/go-micloud/releases/download/1.3/micloud_x86_linux
#国内用户镜像加速替换下面链接中的一个  
#https://github.wuyanzheshui.workers.dev/wangbjun/go-micloud/releases/download/1.3/micloud_x86_linux
#https://download.fastgit.org/wangbjun/go-micloud/releases/download/1.3/micloud_x86_linux
#https://github.91chifun.workers.dev//https://github.com/wangbjun/go-micloud/releases/download/1.3/micloud_x86_linux
# 更改一个你喜欢的名字，我用的是mi,你也可以用其他名字，比如华为之类
sudo mv micloud_x86_linux /usr/local/bin/mi
# 更改权限
sudo chmod u+x /usr/local/bin/mi
# 输入命令确认
mi
# 如果成功会提示你输入账号和密码
```


## 主要功能
- 登录采用模拟小米云服务Web端的登录逻辑，首次登录需输入账号、密码和手机验证码，一次登录，终身有效
- 账号密码本地会暂存，密码已加密，默认位于用户目录下 .micloud.json，长时间不登录无需再输入账号密码，如需完全退出账号，可以删除该配置文件
- 命令支持tab自动补全
- 独家特色分享功能，可以生成一个对外公开分享的链接，让非登录用户快速下载
- 上传和下载采用异步方式，默认上传并发5，下载并发20

## 命令介绍


### 一、login
登录小米云服务账号

所有操作都需要登录小米云服务账号之后才可以进行，所以第一步就是登录，运行本程序需要输入账号密码，以及首次登录需要的手机验证码。

第一次登录之后，程序会保存账号和密码，所以再次运行会尝试采用已经缓存的账号密码登录，如果失败则会要求重新输入账号密码。程序会保存账号密码，不过请放心，密码是以加密的形式保存 **～/.micloud.json** 配置里面，并且该工具绝对不会上传用户账号和密码，如果不放心的话可以查看源码。

如果在登录之后运行 login 命令则会重新登录账号，本程序目前暂不支持多账号切换。

### 二、ls
列出当前目录下的文件
```
ls <dir>
```
这个命令有点类似Linux系统下ls，只不过功能简单，没有额外参数，会显示文件类型、文件大小、创建时间、以及文件名。
```
total 13
d | ------ | 2018-10-26 12:51:07 | Doc
d | ------ | 2018-10-26 13:01:12 | Books
d | ------ | 2018-10-26 13:03:04 | Picture
d | ------ | 2018-10-26 13:03:10 | Package
- | 71 MB  | 2019-12-23 17:20:11 | Geekbench-4.2.3-Linux.tar.gz
- | 69 MB  | 2019-12-23 17:20:56 | Postman-linux-x64-7.1.1.tar.gz
- | 259 MB | 2019-12-23 17:28:36 | wps-office_11.1.0.8722_amd64.deb
- | 492 kB | 2019-12-24 13:49:09 | Baidu_Voice_RestApi_SampleCode.zip
- | 1.0 GB | 2019-12-24 13:53:30 | Deepin-Apps-Installation.zip
```

### 三、cd
切换到目录
```
cd <dir>
```
这个命令类似Linux下的cd，但是功能有限，只支持进入目录和退出目录。

### 四、download
下载当前目录下的一个文件或者文件夹，如果是文件夹则会递归下载文件夹里面的所有文件
```
download <file> [-d dir]
```
下载文件存在位置默认是当前工具的运行目录，支持指定下载位置，通过 -d 参数可以传递一个目录
```
Go@MiCloud:/Books$ download Go高级编程.pdf
2020-09-26 23:28:50 #添加下载任务: /Go高级编程.pdf
```
指定下载位置：
```
Go@MiCloud:/Books$ download Go高级编程.pdf -d /home/jwang/Downloads
2020-09-26 23:31:52 #添加下载任务: /home/jwang/Downloads/Go高级编程.pdf
```

### 五、upload
上传一个文件或者文件夹到当前所在目录，如果是文件夹则会递归上传目录里面的所有文件，需要注意的是路径必须是绝对路径，如 /home/jwang/abc.jpg。

由于小米云服务服务端的限制，目前支持单个文件最大4GB，超过这个大小无法上传。
```
upload <filepath>
```
上传单个文件：
```
Go@MiCloud:/Books$ upload /home/jwang/Downloads/Go高级编程.pdf
2020-09-26 23:33:02 #添加上传任务: /home/jwang/Downloads/Go高级编程.pdf
```
上传整个目录：
```
Go@MiCloud:/Books$ upload /home/jwang/Books
2020-09-26 23:33:45 #添加上传任务: /home/jwang/Books/Clean Code-代码整洁之道.pdf
2020-09-26 23:33:45 #添加上传任务: /home/jwang/Books/Go高级编程.pdf
2020-09-26 23:33:45 #添加上传任务: /home/jwang/Books/Laravel框架关键技术解析-陈昊.pdf
2020-09-26 23:33:45 #添加上传任务: /home/jwang/Books/Linux就该这么学.pdf
2020-09-26 23:33:45 #添加上传任务: /home/jwang/Books/Linux程序设计 中文第4版.pdf
2020-09-26 23:33:45 #添加上传任务: /home/jwang/Books/MySQL性能调优与架构设计--全册.pdf
...
...
...
```

### 六、share
生成一个对外公开分享的链接，可以提供给未登录用户下载
```
share <file>
```
理论上说小米网盘的文件只能登录自己的小米账号自己下载，但是通过接口可以获取一个非公开的链接，提供给其它用户下载，这里采用了一个短网址，链接有效期是24小时，不过下载速度非常快，也不用开会员。
```
Go@MiCloud:~$ share wps-office_11.1.0.8722_amd64.deb
获取分享链接成功(采用了短链接，有效期24小时): http://t.wibliss.com/BRfnl
```

### 七、rm
删除文件或者目录，实际上是放入回收站，如果想真正删除，可以登录小米云服务官方网页端，在回收站里面再删一次。
```
Go@MiCloud:/Books$ rm Books
2020-09-26 23:38:37 #[ Books ]删除成功
```

### 八、mkdir
在当前目录下创建目录，功能比较简单，无其它参数
```
Go@MiCloud:$ ls
total 6
d | ------ | 2018-10-26 12:51:07 | Doc
d | ------ | 2018-10-26 13:01:12 | Books
d | ------ | 2018-10-26 13:03:04 | Picture
d | ------ | 2018-10-26 13:03:10 | Package
d | ------ | 2020-08-22 16:58:31 | Movies
- | 1.9 kB | 2020-09-21 20:54:50 | main.go
Go@MiCloud:$ rm main.go
2020-09-21 20:54:53 #[ main.go ]删除成功
Go@MiCloud:$ mkdir 2222
2020-09-21 20:55:07 #[ 2222 ]创建成功
Go@MiCloud:$ ls
total 6
d | ------ | 2018-10-26 12:51:07 | Doc
d | ------ | 2018-10-26 13:01:12 | Books
d | ------ | 2018-10-26 13:03:04 | Picture
d | ------ | 2018-10-26 13:03:10 | Package
d | ------ | 2020-08-22 16:58:31 | Movies
d | ------ | 2020-09-21 20:55:07 | 2222
Go@MiCloud:$ 
```
### 九、tree
显示当前目录树结构，但是此命令只会展示曾经进入过的目录，功能有限，仅供参考。
```
/
├── Doc
├── Books
├── Picture
│   ├── 10-11.jpg
│   ├── 10-10.jpg
│   ├── 10-13-beta.jpg
│   ├── 10-9.jpg
│   ├── 1493959153401.jpg
│   ├── 1493959163558.jpg
│   ├── 1493959172685.jpg
│   ├── 1493959186771.jpg
│   ├── 1493959197414.jpg
│   ├── wallPaper
│   │   ├── Solid Colors
│   │   ├── 1F51Q05K0-2.jpg
│   │   ├── 150505104113-9.jpg
│   │   ├── 150605101120-17.jpg
│   │   ├── 1F505102532-11.jpg
│   │   ├── 1493959163558.jpg
│   │   ├── 1493959186771.jpg
│   │   ├── 1493959197414.jpg
│   │   ├── El Capitan.jpg
│   │   ├── Elephant.jpg
│   │   ├── Flamingos.jpg
│   │   ├── Floating Ice.jpg
│   │   ├── Floating Leaves.jpg
├── Package
...
...
...
```

### 十、jobs
目前上传和下载都是异步的，jobs命令可以显示上传和下载的任务信息列表，便于查看上传和下载结果以及进度。
```
Go@MiCloud:$ download WeGameMiniLoader.3.25.1.8081.gw.exe
2020-09-22 00:26:37 #添加下载任务: /WeGameMiniLoader.3.25.1.8081.gw.exe
Go@MiCloud:$ download navicat15-premium-en.AppImage
2020-09-22 00:26:39 #添加下载任务: /navicat15-premium-en.AppImage
Go@MiCloud:$ download phpStudy.zip
2020-09-22 00:26:42 #添加下载任务: /phpStudy.zip
Go@MiCloud:$ download main.go
2020-09-22 00:26:44 #添加下载任务: /main.go
Go@MiCloud:$ jobs
--------------------------------------------------------------------------------
任务状态 |状态信息 |文件总大小 |已处理大小 |文件名
--------------------------------------------------------------------------------
已完成   |下载成功     |34 MB    |34 MB    |phpStudy.zip
已完成   |下载成功     |4.3 MB   |4.3 MB   |WeGameMiniLoader.3.25.1.8081.gw.exe
已完成   |下载成功     |1.9 kB   |1.9 kB   |main.go
--------------------------------------------------------------------------------
下载中   |正在下载     |140 MB   |80 MB    |navicat15-premium-en.AppImage
--------------------------------------------------------------------------------
总任务 4 个，已完成 3 个, 待处理 0 个，处理中 1 个
```
也可以下载文件夹，会批量创建任务：
```
Go@MiCloud:$ jobs
--------------------------------------------------------------------------------
任务状态 |状态信息   |文件总大小 |已处理大小 |文件名
--------------------------------------------------------------------------
已完成   |下载成功   |7.7 MB     |7.7 MB  |Vim实用技巧.pdf
已完成   |下载成功   |6.6 MB     |6.6 MB  |Go高级编程.pdf
已完成   |下载成功   |3.2 MB     |3.2 MB  |MySQL性能调优与架构设计--全册.pdf
已完成   |下载成功   |660 kB     |660 kB  |easy-swoole.pdf
已完成   |上传成功   |1.9 kB     |1.9 kB  |main.go
-----------------------------------------------------------------------
下载中   |正在下载   |105 MB     |22 MB   |Linux程序设计 中文第4版.pdf
下载中   |正在下载   |82 MB      |42 MB   |Laravel框架关键技术解析-陈昊.pdf
下载中   |正在下载   |80 MB      |4.7 MB  |啊哈！算法.pdf
下载中   |正在下载   |16 MB      |704 kB  |Clean Code-代码整洁之道.pdf
下载中   |正在下载   |1.2 MB     |262 kB  |学习+Go+语言(Golang).pdf
-----------------------------------------------------------------------
待处理   |等待下载   |525 MB     |0 B     |深入理解计算机系统（原书第三版）.pdf
待处理   |等待下载   |79 MB      |0 B     |汇编语言(第3版)王爽著.pdf
待处理   |等待下载   |2.0 MB     |0 B     |深入理解PHP内核.pdf
待处理   |等待下载   |1.2 MB     |0 B     |阿里巴巴Java开发手册（详尽版）.pdf
待处理   |等待下载   |108 kB     |0 B     |前端开发工程师-4年-夏恒.pdf
--------------------------------------------------------------------------------
总任务 25 个，已完成 15 个, 待处理 0 个，处理中 10 个
```
### 十一、quit（exit）
退出程序，也可以通过快捷键 Ctrl+C、Ctrl+D进行操作。

### 十二、lsAlbum
列出所有相册，显示相册里面文件数和更新时间
```
Go@MiCloud:$ lsAlbum 
total 6
文件数 |    最后更新时间     | 相册名
---------------------------
5530   | 2020-09-26 13:09:05 | 相机
22     | 2020-08-01 18:22:55 | 私密
6      | 2020-03-09 12:21:19 | 我的创作
8      | 2020-08-18 19:42:18 | 屏幕录制
2929   | 2020-09-25 20:55:49 | 截屏
2975   | 2020-04-04 18:19:54 | 历史照片
```

### 十三、downloadAlbum
下载相册里面照片，参数是相册名，如果无参数则会下载所有相册及其里面所有照片。
```
downloadAlbum 相机
```
其下载逻辑和文件下载类似，但是只支持下载整个相册，不支持单个照片。
