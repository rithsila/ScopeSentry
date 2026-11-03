# ScopeSentry (Headless & Agent-First)

中文 | [English](./README.md)

## 介绍
Scope Sentry是一款具有分布式资产测绘、子域名枚举、信息泄露检测、漏洞扫描、目录扫描、子域名接管、爬虫、页面监控功能的工具，通过构建多个节点，自由选择节点运行扫描任务。当出现新漏洞时可以快速排查关注资产是否存在相关组件。

分布式搭建参考文章：[https://mp.weixin.qq.com/s/xfgRxUjljoQ8KzacblktxA](https://mp.weixin.qq.com/s/xfgRxUjljoQ8KzacblktxA)

服务器推荐：[lightnode](https://www.lightnode.com/?inviteCode=CQ11JU&promoteWay=LINK)

## 架构
- **服务端**: Go (Gin) Headless 服务端，原生集成 Model Context Protocol (MCP) 与 REST API
- **CLI 客户端**: 独立轻量级 Go CLI 客户端 (`scopesentry`)
- **扫描端**: Go 分布式扫描节点 (`ScopeSentry-Scan`)
- **数据库**: MongoDB & Redis

## Agent 与自动化支持
ScopeSentry 专为 AI Agent 和自动化攻击面发现而设计：
- **内置 MCP 服务**: 在 `/mcp` 端点暴露全部 17 个安全扫描和资产管理工具。
- **独立 CLI**: `scopesentry` 纯 JSON stdout 输出，便于管道与脚本集成。
- **无头凭证管理**: 通过 `./ScopeSentry apikey create --name <name>` 生成 API Key。
- **开发指南**: 参阅 [AI_AGENT_GUIDE.md](./AI_AGENT_GUIDE.md) 与 [SKILL.md](./SKILL.md)。

## 网址

- 官网&安装教程：[https://www.scope-sentry.top](https://www.scope-sentry.top)
- Github: [https://github.com/Autumn-27/ScopeSentry](https://github.com/Autumn-27/ScopeSentry)
- 扫描端源码：[https://github.com/Autumn-27/ScopeSentry-Scan](https://github.com/Autumn-27/ScopeSentry-Scan)
- 前端源码：[https://github.com/Autumn-27/ScopeSentry-UI](https://github.com/Autumn-27/ScopeSentry-UI)
- 插件市场: [插件市场](https://plugin.scope-sentry.top/)
- 插件模板：[https://github.com/Autumn-27/ScopeSentry-Plugin-Template](https://github.com/Autumn-27/ScopeSentry-Plugin-Template)


## 安装
```
git clone https://github.com/Autumn-27/ScopeSentry.git
cd ScopeSentry
# 修改.env文件中的mongodb、redis账号密码。
docker-compose -f single-host-deployment.yml up -d
```
运行之后一共四个容器scope-sentry (服务端-web界面)、scopesentry-scan (扫描端)、Mongodb (数据库)、Redis。（默认会有一个扫描节点）

查看初始用户密码以及插件二次验证密码

```docker logs scope-sentry```


**添加新的节点(可选)**
```
git clone https://github.com/Autumn-27/ScopeSentry-Scan.git
cd ScopeSentry-Scan/build
# 编辑 .env文件中Mongdb、Redis的连接信息，NodeName为节点名称，每个节点名称请保持唯一性（如果为空会随机生成，可在web界面更改名称）。
docker-compose -f scan-docker-compose.yml up -d
```


## 插件流程图

<img src="流程图.svg"/>


## 目前功能
- 插件系统（通过拓展的方式加入任何工具）
- 子域名枚举
- 子域名接管检测
- 端口扫描
- ICP自动化收集
- APP自动化收集
- 小程序自动化收集
- 资产识别
- 目录扫描
- 漏洞扫描
- 敏感信息泄露检测(支持扫描pdf)
- URL提取
- 爬虫
- 页面监控
- 自定义WEB指纹
- POC导入
- 资产分组
- 多节点扫描
- webhook
- 数据导出

## To DO
- 弱口令爆破
- 关系图

## 安装

安装教程见[官网](https://www.scope-sentry.top)

## 交流
联系方式见下方

## 命令行与 MCP 使用

ScopeSentry 作为无头扫描系统，通过 Model Context Protocol (MCP) 和专属 CLI 工具进行交互。

### 健康检查与状态诊断
```bash
scopesentry doctor
```

### 无头管理凭证
```bash
# 创建管理员 API Key
./ScopeSentry apikey create --name admin

# 查看 API Key 列表
./ScopeSentry apikey list

# 获取插件密钥
./ScopeSentry pluginkey
```

### CLI 资产与任务查询
```bash
# 查询项目列表
scopesentry project list

# 查询特定资产 (支持语法DSL)
scopesentry asset list --type subdomain --search 'project == "default"'

# 创建扫描项目
scopesentry project create --name "example" --target "example.com"
```

# 许可证
该项目所有分支遵循AGPL-3.0，另外需要遵循附加条款：
1. 本软件的商业用途需要单独的商业许可。
2. 公司、组织和营利性实体在使用、分发或修改本软件之前必须获得商业许可。
3. 个人和非营利组织可以根据 AGPL-3.0 的条款自由使用本软件。
4. 如有商业许可查询，请联系 rainy-autumn@outlook.com。
