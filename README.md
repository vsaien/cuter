
# 目录介绍:
	1. application 应用程序目录
	2. common  公共模块目录
	3. library 相关库服务目录
	4. example 代码示例目录

例如：现开发CRM系统相关的接口<br>
开发规范如下
需要在application 目录下创建crm 目录<br>

开发crm系统中user服务接口时<br>
在上一个步骤application/crm 目录下创建user目录<br>

user目录结构为<br>

	cmd    // 可执行文件目录
    	api  // api 接口启动入口目录
			config
				config.go  用来解析配置.json的结构体
				config.json 配置json
		
    	rpc  // rpc 启动入口目录
			config
				config.go  用来解析配置.json的结构体
				config.json 配置json
		
	handler // 处理方法目录集
	logic // 业务逻辑方法目录集
	model // model集
	rpcproto  // proto文件集
	rpcserver // 关于此服务提供商rpc服务文件集






