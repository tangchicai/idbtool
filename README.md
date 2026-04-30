```md
# LevelDB Dynamic Viewer

一个用于查看带有 **自定义 comparer** 的 LevelDB 数据库的小工具。
A small tool for inspecting LevelDB databases that use a **custom comparer**.

---

## Features / 功能

- **Dynamic comparer retry**
  自动识别 comparer 名称，并在打开失败时重试。
  Automatically detects the comparer name and retries when opening fails.

- **Search by key**
  使用 `-search` 精确查询某个 key。
  Use `-search` to look up a specific key.

- **Prefix scan**
  使用 `-prefix` 查看指定前缀下的所有 key/value。
  Use `-prefix` to list all key/value pairs under a given prefix.

- **Full scan**
  使用 `-scan` 遍历整个数据库。
  Use `-scan` to iterate through the entire database.

---

## Build / 编译

### Windows
```powershell
go build -o leveldb-viewer.exe
```

### Linux / macOS
```bash
go build -o leveldb-viewer
```

---

## Usage / 使用方法

### 1. Search a key / 精确查找
```bash
leveldb-viewer -db "C:\Users\Administrator\Downloads\123\leveldb" -search "20260429"
```

### 2. Prefix scan / 前缀扫描
```bash
leveldb-viewer -db "C:\Users\Administrator\Downloads\123\leveldb" -prefix "2026"
```

### 3. Full scan / 全库扫描
```bash
leveldb-viewer -db "C:\Users\Administrator\Downloads\123\leveldb" -scan
```

---

## How it works / 工作原理

When opening a LevelDB database, you may see an error like:

打开 LevelDB 数据库时，可能会遇到类似错误：

```text
comparer mismatch: use idb_cmp1, but got 'BytewiseComparator'
```

This tool will:

这个工具会：

1. Parse the comparer name from the error message
   从错误信息中解析 comparer 名称

2. Ask whether to retry with that comparer
   询问是否使用该 comparer 重新打开

3. Reopen the database dynamically
   使用动态 comparer 重新打开数据库

---

## Example Output / 输出示例

### Search / 查找
```text
DB opened successfully
key="20260429"  value=68656c6c6f
```

### Scan / 扫描
```text
DB opened successfully
key="key1"  value=313233
key="key2"  value=616263
```

---

## Notes / 注意事项

- The comparer name is handled dynamically, but the actual comparison logic is simplified.
  comparer 名称是动态处理的，但实际比较逻辑是简化版。

- If the original database uses a non-standard sort order, some queries may not behave exactly like the original application.
  如果原数据库使用了非标准排序规则，某些查询行为可能与原应用不完全一致。

- For best results, use the same comparer behavior as the original database creator.
  为了获得最佳效果，建议尽量使用与原数据库创建时相同的 comparer 行为。

---

---

## Contributing / 贡献

Pull requests and issues are welcome.
欢迎提交 Pull Request 和 Issue。
```
