### 依赖基础组件
- redis
- etcd
- git仓库
- consul

### 部署配置
- 部署 consul ，利用consul来完成服务注册，发现，健康检测
- 部署 Redis,etcd,MySQL。
- 新建git repo，可以参考 https://github.com/taolx0/config-service 创建对应项目的文件，修改Redis，MySQL，etcd等组件的配置
- 使用spring-cloud-config部署Config-Service配置文件

### 测试
- 使用postman工具创建活动
- 使用jmeter工具进行并发测试


