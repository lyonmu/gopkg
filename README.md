# gopkg

一个 Go 语言工具包，收集和整理在开发工作中常用的工具函数。

## 项目结构

```
gopkg/
├── pinyin/    # 中文拼音转换工具
├── version/   # 版本信息管理
├── id/        # 分布式唯一 ID 生成器
├── structx/   # struct 操作工具
└── ...        # 持续收集中...
```

## 模块

| 模块 | 说明 |
|------|------|
| [pinyin](pinyin/README.md) | 中文拼音转换工具，支持首字母、完整拼音、驼峰格式及中文检测 |
| [version](version/README.md) | 版本信息管理，支持编译时注入 |
| [id](id/README.md) | 分布式唯一 ID 生成器，基于 Sonyflake 算法 |
| [structx](structx/README.md) | struct 操作工具，支持转 map、差异比较、属性赋值 |

## 安装

```bash
go get github.com/lyonmu/gopkg
```

## License

MIT License. 详见 [LICENSE](LICENSE) 文件。
