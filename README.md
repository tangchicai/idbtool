# IndexedDB LevelDB Search & Delete Tool
一个用于搜索和删除 Chromium/Chrome 系列浏览器 `IndexedDB` 底层 `LevelDB` 数据的 Go 工具。
## 功能特性
- 支持扫描 LevelDB 数据
- 同时搜索 `key` 和 `value`
- 支持模糊匹配
- 自动规范化文本，提升命中率
- 可匹配普通字符串、宽字符痕迹、隔字符内容
- 交互式删除
  - 删除单条
  - 删除多条
  - 删除全部匹配项
- 适合排查浏览器 IndexedDB 本地数据
## 使用场景
适用于以下场景：
- 查找浏览器 IndexedDB 中的特定时间、ID、字符串
- 定位某个业务数据是否还存在
- 清理错误或脏数据
- 分析 Chromium 系应用本地存储内容
## 原理说明
本工具直接打开 IndexedDB 底层的 LevelDB 目录，遍历所有 key/value，并对内容做以下处理：
1. 原始字符串匹配
2. 去除非字母数字后的规范化匹配
3. 从二进制内容中提取可见字符串后再次匹配
这样可以提高对以下数据的命中率：
- 普通字符串
- UTF-16 / 宽字符痕迹
- 二进制中夹杂的可见文本
- 被隔字符包裹的日期/时间字符串
## 环境要求
- Go 1.20+（建议）
- Windows / Linux / macOS
- 目标 LevelDB 目录可访问
## 安装
```bash
git clone https://github.com/tangchicai/idbtool.git
cd indexeddb-leveldb-tool
go mod tidy
go build -o idbtool
