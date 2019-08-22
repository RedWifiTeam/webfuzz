# WebFuzz

## DirFuzz

随便写的一个迭代目录爆破工具

~~懒得写说明了~~

**简单说明**

```bash
# Build
go mod download
go build

# Add Dicts
./webfuzz d -j -i dicts/dic.txt

# Fuzz
./webfuzz f -u http://www.xxx.com

# Others
./webfuzz -h
```