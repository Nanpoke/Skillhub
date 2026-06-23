# SkillHub - Go 语言快速入门指南

## 为什么要用 Go？

Go 语言非常适合桌面工具开发：
- ✅ **语法简单** - 比 C++ 简单，比 Python 快
- ✅ **编译成单个可执行文件** - 用户无需安装依赖
- ✅ **文件系统操作强大** - 完美适合 SkillHub 的需求
- ✅ **跨平台** - Windows/Mac/Linux 都能编译

---

## 15 分钟 Go 速成

### 1. 基础语法（和 JavaScript 对比）

```go
// JavaScript
const name = "SkillHub";
let count = 0;
function add(a, b) {
    return a + b;
}

// Go
var name = "SkillHub"  // 或 name := "SkillHub" (简短声明)
count := 0             // 自动推断类型
func add(a int, b int) int {
    return a + b
}
```

### 2. 变量和类型

```go
// Go 是强类型语言，但类型推断很方便
name := "SkillHub"           // string
version := 1.0               // float64
count := 42                  // int
isReady := true              // bool

// 显式声明
var description string = "AI Skill Manager"
var skills []string          // 字符串数组（切片）
```

### 3. 结构体（类似 JavaScript 对象）

```go
// 定义结构体（类似于 TypeScript interface）
type Skill struct {
    Name    string
    Version string
    Tags    []string
}

// 创建实例
skill := Skill{
    Name:    "frontend-design",
    Version: "1.0.0",
    Tags:    []string{"UI", "UX"},
}

// 访问字段
fmt.Println(skill.Name)  // "frontend-design"
```

### 4. 函数和方法

```go
// 普通函数
func installSkill(url string) error {
    // 实现安装逻辑
    return nil  // nil 表示没有错误
}

// 方法（给结构体绑定函数）
func (s *Skill) Enable() error {
    s.Enabled = true
    return nil
}

// 使用
skill := Skill{Name: "test"}
skill.Enable()  // 调用方法
```

### 5. 错误处理（Go 的特色）

```go
// Go 没有 try-catch，用多返回值处理错误
func readFile(path string) (string, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return "", err  // 返回错误
    }
    return string(content), nil  // nil 表示成功
}

// 使用
content, err := readFile("config.json")
if err != nil {
    fmt.Println("读文件失败:", err)
    return
}
fmt.Println(content)
```

### 6. 文件操作（SkillHub 核心）

```go
import "os"

// 检查文件是否存在
func fileExists(path string) bool {
    _, err := os.Stat(path)
    return !os.IsNotExist(err)
}

// 读取文件
content, err := os.ReadFile("skill.md")

// 写入文件
err := os.WriteFile("config.json", data, 0644)

// 创建目录
err := os.MkdirAll("~/.skill-hub/skills", 0755)

// 复制文件
import "io"
func copyFile(src, dst string) error {
    input, err := os.ReadFile(src)
    if err != nil {
        return err
    }
    return os.WriteFile(dst, input, 0644)
}
```

### 7. JSON 处理（配置读写）

```go
import "encoding/json"

// 结构体转 JSON
config := Config{Theme: "dark"}
data, err := json.MarshalIndent(config, "", "  ")
// 结果: {"theme": "dark"}

// JSON 转结构体
var config Config
err := json.Unmarshal(data, &config)

// 读写文件
func saveConfig(config Config) error {
    data, _ := json.MarshalIndent(config, "", "  ")
    return os.WriteFile("config.json", data, 0644)
}
```

---

## SkillHub 核心概念映射

| JavaScript/前端概念 | Go 等价物 |
|-------------------|----------|
| `Object` | `struct` |
| `Array` | `slice` (`[]Type`) |
| `Promise/async-await` | 直接返回，或 Go 的 `goroutine` |
| `JSON.parse/stringify` | `json.Unmarshal/Marshal` |
| `require('fs')` | `os` 包 |
| `console.log` | `fmt.Println` |
| `null/undefined` | `nil` |
| `try-catch` | 多返回值 `(result, error)` |

---

## Wails 特有概念

### 前端调用后端（绑定）

```go
// backend/app.go
type App struct {
    ctx context.Context
}

// 给前端调用的方法
func (a *App) GetSkills() []Skill {
    // 返回 Skill 列表
    return skills
}

func (a *App) InstallSkill(url string) error {
    // 安装逻辑
    return nil
}
```

```javascript
// 前端调用（Wails 自动生成绑定）
import { GetSkills, InstallSkill } from '../wailsjs/go/main/App';

// 直接调用 Go 函数！
const skills = await GetSkills();
await InstallSkill("https://github.com/...");
```

---

## 开发时你会用到的 Go 代码模板

### 读取配置文件

```go
func LoadConfig() (*Config, error) {
    data, err := os.ReadFile("config.json")
    if err != nil {
        return nil, err
    }
    
    var config Config
    err = json.Unmarshal(data, &config)
    return &config, err
}
```

### 遍历目录

```go
func ListSkills() ([]string, error) {
    entries, err := os.ReadDir("~/.skill-hub/skills")
    if err != nil {
        return nil, err
    }
    
    var skills []string
    for _, entry := range entries {
        if entry.IsDir() {
            skills = append(skills, entry.Name())
        }
    }
    return skills, nil
}
```

### HTTP 请求（爬取 skills.sh）

```go
import "net/http"
import "io"

func FetchSkillsSh() (string, error) {
    resp, err := http.Get("https://skills.sh/api/skills")
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    return string(body), nil
}
```

---

## 学习资源

1. **Go 官方教程**（30分钟）: https://go.dev/doc/tutorial/getting-started
2. **Wails 文档**: https://wails.io/docs/introduction
3. **Go by Example**（交互式学习）: https://gobyexample.com/

---

**不需要成为 Go 专家！** 你只需要理解上面的基础语法，然后复制我提供的代码模板即可。90% 的代码我会为你写好！🚀
