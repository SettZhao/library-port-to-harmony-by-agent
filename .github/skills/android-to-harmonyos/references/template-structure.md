# Template 项目结构指南

本文档说明如何使用 OpenHarmony 标准项目模板（Template）组织移植后的代码。

---

## Template 项目概览

Template 是 OpenHarmony 的标准项目结构，包含应用入口、库模块、原生代码和测试：

```
Template/
├── AppScope/                    # 应用全局配置
│   └── resources/              # 全局资源文件
├── cpp/                        # Native C++ 代码
│   ├── CMakeLists.txt          # CMake 构建配置
│   ├── napi_init.cpp           # NAPI 入口
│   └── types/                  # Native 类型定义
│       └── libentry/
│           ├── index.d.ts      # TypeScript 类型声明
│           └── oh-package.json5
├── entry/                      # 应用入口模块 (HAP)
│   ├── src/main/               # 主代码
│   │   ├── ets/
│   │   │   ├── entryability/   # Ability 入口
│   │   │   └── pages/          # UI 页面
│   │   └── resources/          # 资源文件
│   ├── src/ohosTest/           # 测试代码
│   │   └── ets/test/
│   ├── build-profile.json5     # 模块构建配置
│   └── oh-package.json5        # 模块依赖
├── library/                    # 库模块 (HAR)
│   ├── src/main/
│   │   ├── ets/                # ArkTS 源码
│   │   │   └── index.ets       # 库导出入口
│   │   └── resources/          # 库资源
│   ├── Index.ets               # 对外导出索引
│   ├── build-profile.json5
│   └── oh-package.json5
├── hvigorfile.ts               # 项目构建脚本
├── build-profile.json5         # 项目构建配置
└── oh-package.json5            # 项目依赖
```

---

## Android 库到 Template 的映射

### 核心模块映射

| Android 模块 | Template 模块 | 说明 |
|-------------|--------------|------|
| `mylibrary/` (Library Module) | `library/` | 库代码主目录 |
| `app/` (Sample App) | `entry/` | 示例应用/入口 |
| `mylibrary/src/main/java/` | `library/src/main/ets/` | 库源码 (Java → ArkTS) |
| `mylibrary/src/androidTest/` | `entry/src/ohosTest/` | 测试代码 |
| `mylibrary/src/main/cpp/` | `cpp/` | Native C++ 代码 |
| `app/src/main/java/` | `entry/src/main/ets/` | 示例代码 |

### 文件映射示例

```
Android 库结构                      → Template 结构
─────────────────────────────────────────────────────────
mylibrary/
├── src/main/java/
│   └── com/example/mylib/
│       ├── MyClass.java           → library/src/main/ets/MyClass.ets
│       ├── utils/
│       │   └── Helper.java        → library/src/main/ets/utils/Helper.ets
│       └── api/
│           └── ApiClient.java     → library/src/main/ets/api/ApiClient.ets
│
├── src/main/cpp/
│   ├── CMakeLists.txt             → cpp/CMakeLists.txt
│   └── native-lib.cpp             → cpp/napi_init.cpp
│
├── src/androidTest/
│   └── MyClassTest.java           → entry/src/ohosTest/ets/test/MyClass.test.ets
│
└── build.gradle                   → library/build-profile.json5

app/
└── src/main/java/
    └── com/example/app/
        └── MainActivity.java      → entry/src/main/ets/pages/Index.ets
```

---

## library/ 模块详解

library/ 目录是移植后的 OpenHarmony 库的主体。

### 目录结构

```
library/
├── src/main/
│   ├── ets/                       # ArkTS 源码
│   │   ├── components/           # UI 组件
│   │   ├── models/               # 数据模型
│   │   ├── utils/                # 工具类
│   │   ├── api/                  # API 接口
│   │   └── index.ets             # 内部导出
│   └── resources/                # 资源文件
│       ├── base/
│       │   ├── element/          # 字符串等资源
│       │   └── media/            # 图片资源
│       └── rawfile/              # 原始文件
├── Index.ets                      # 对外导出索引 ⭐
├── build-profile.json5            # 构建配置
└── oh-package.json5               # 依赖配置
```

### Index.ets 导出示例

```typescript
// library/Index.ets - 对外导出的公共 API
export { MyClass } from './src/main/ets/MyClass';
export { Helper } from './src/main/ets/utils/Helper';
export { ApiClient } from './src/main/ets/api/ApiClient';
export { MyComponent } from './src/main/ets/components/MyComponent';

// 类型导出
export type { Config } from './src/main/ets/models/Config';
export type { Result } from './src/main/ets/models/Result';
```

### build-profile.json5

```json5
{
  "apiType": "stageMode",
  "buildOption": {
    "arkOptions": {
      "runtimeOnly": {
        "sources": [
          "./src/main/ets"
        ]
      }
    }
  },
  "targets": [
    {
      "name": "default",
      "runtimeOS": "HarmonyOS"
    }
  ]
}
```

### oh-package.json5

```json5
{
  "name": "library",          // 库名称
  "version": "1.0.0",
  "description": "My OpenHarmony Library",
  "main": "Index.ets",        // 入口文件
  "author": "",
  "license": "Apache-2.0",
  "dependencies": {
    // 运行时依赖
  },
  "devDependencies": {
    // 开发依赖
    "@ohos/hypium": "^1.0.4"
  }
}
```

---

## entry/ 模块详解

entry/ 目录是示例应用，展示如何使用移植后的库。

### 目录结构

```
entry/
├── src/main/
│   ├── ets/
│   │   ├── entryability/         # Ability 入口
│   │   │   └── EntryAbility.ets
│   │   └── pages/                # UI 页面
│   │       └── Index.ets         # 主页面
│   ├── resources/                # 应用资源
│   └── module.json5              # 模块配置
├── src/ohosTest/                 # 测试代码 ⭐
│   └── ets/
│       ├── test/
│       │   ├── Ability.test.ets  # 测试入口
│       │   └── List.test.ets     # 测试列表
│       └── testability/
│           └── TestAbility.ets   # 测试 Ability
├── build-profile.json5
└── oh-package.json5
```

### entry/oh-package.json5

```json5
{
  "name": "entry",
  "version": "1.0.0",
  "description": "Sample app using library",
  "main": "",
  "author": "",
  "license": "",
  "dependencies": {
    "library": "file:../library"    // 引用本地 library 模块 ⭐
  },
  "devDependencies": {
    "@ohos/hypium": "^1.0.4"
  }
}
```

### 使用库 API

```typescript
// entry/src/main/ets/pages/Index.ets
import { MyClass, Helper } from 'library';  // 导入库

@Entry
@Component
struct Index {
  @State message: string = '';

  aboutToAppear() {
    // 使用库 API
    let obj = new MyClass();
    this.message = obj.getMessage();
  }

  build() {
    Column() {
      Text(this.message)
        .fontSize(20)
    }
  }
}
```

---

## cpp/ 模块详解 (Native 代码)

如果 Android 库包含 JNI/NDK 代码，需要迁移到 cpp/ 目录。

### 目录结构

```
cpp/
├── CMakeLists.txt                 # CMake 构建配置
├── napi_init.cpp                  # NAPI 入口注册
├── hello.cpp                      # Native 实现
├── hello.h                        # 头文件
└── types/libentry/                # TypeScript 类型声明
    ├── index.d.ts                 # 类型定义 ⭐
    └── oh-package.json5
```

### napi_init.cpp

```cpp
#include "napi/native_api.h"
#include "hello.h"

// 注册 Native 函数
static napi_value Add(napi_env env, napi_callback_info info) {
    // 实现...
}

// 模块初始化
EXTERN_C_START
static napi_value Init(napi_env env, napi_value exports) {
    napi_property_descriptor desc[] = {
        {"add", nullptr, Add, nullptr, nullptr, nullptr, napi_default, nullptr}
    };
    napi_define_properties(env, exports, sizeof(desc) / sizeof(desc[0]), desc);
    return exports;
}
EXTERN_C_END

// 模块定义
static napi_module demoModule = {
    .nm_version = 1,
    .nm_flags = 0,
    .nm_filename = nullptr,
    .nm_register_func = Init,
    .nm_modname = "entry",  // 模块名
    .nm_priv = nullptr,
    .reserved = {0},
};

extern "C" __attribute__((constructor)) void RegisterEntryModule(void) {
    napi_module_register(&demoModule);
}
```

### types/libentry/index.d.ts

```typescript
// Native 模块的 TypeScript 类型声明
export const add: (a: number, b: number) => number;
export const multiply: (a: number, b: number) => number;
```

### CMakeLists.txt

```cmake
cmake_minimum_required(VERSION 3.5.0)
project(mylib)

# 添加源文件
add_library(entry SHARED
    napi_init.cpp
    hello.cpp
)

# 链接系统库
target_link_libraries(entry PUBLIC
    libace_napi.z.so
    libhilog_ndk.z.so
)
```

---

## 测试代码组织

测试代码统一放在 entry/src/ohosTest/ 目录。

### 测试目录结构

```
entry/src/ohosTest/ets/
├── test/
│   ├── Ability.test.ets           # 测试入口
│   ├── List.test.ets              # 测试套件列表
│   ├── MyClass.test.ets           # 单元测试
│   └── ApiClient.test.ets         # API 测试
└── testability/
    └── TestAbility.ets            # 测试 Ability
```

### Ability.test.ets

```typescript
import { describe, it, expect } from '@ohos/hypium';
import { MyClass } from 'library';

export default function abilityTest() {
  describe('MyClass Tests', () => {
    
    it('shouldCreateInstance', () => {
      let obj = new MyClass();
      expect(obj).assertNotNull();
    });
    
    it('shouldReturnCorrectMessage', () => {
      let obj = new MyClass();
      let msg = obj.getMessage();
      expect(msg).assertEqual('Hello OpenHarmony');
    });
    
  });
}
```

### List.test.ets

```typescript
import abilityTest from './Ability.test';

export default function testsuite() {
  // 注册所有测试套件
  abilityTest();
}
```

---

## 资源文件组织

### library/src/main/resources/

库的资源文件：

```
library/src/main/resources/
├── base/
│   ├── element/
│   │   ├── string.json          # 字符串资源
│   │   └── color.json           # 颜色资源
│   └── media/
│       └── icon.png             # 图片资源
└── rawfile/                     # 原始文件
    └── config.json
```

### string.json 示例

```json
{
  "string": [
    {
      "name": "app_name",
      "value": "My Library"
    },
    {
      "name": "welcome_message",
      "value": "Welcome to OpenHarmony"
    }
  ]
}
```

### 使用资源

```typescript
import resourceManager from '@ohos.resourceManager';

// 获取字符串资源
let str = await this.context.resourceManager.getStringValue($r('app.string.welcome_message'));

// 获取图片资源
Image($r('app.media.icon'))
```

---

## 构建配置文件

### 项目级 build-profile.json5

```json5
{
  "app": {
    "signingConfigs": [],
    "products": [
      {
        "name": "default",
        "signingConfig": "default"
      }
    ]
  },
  "modules": [
    {
      "name": "entry",
      "srcPath": "./entry"
    },
    {
      "name": "library",
      "srcPath": "./library"
    }
  ]
}
```

### 项目级 oh-package.json5

```json5
{
  "name": "myproject",
  "version": "1.0.0",
  "description": "OpenHarmony project",
  "main": "",
  "author": "",
  "license": "",
  "dependencies": {},
  "devDependencies": {}
}
```

---

## 迁移工作流

### 第一步：创建项目结构

```bash
# 在 DevEco Studio 中创建项目时选择 "Library" 模板
# 或手动复制 Template 目录结构
```

### 第二步：迁移核心代码

```
Android Java/Kotlin 代码 → library/src/main/ets/
```

- 将 Android 库的主要代码转换为 ArkTS
- 保持原有的包结构和类名
- 在 library/Index.ets 中导出公共 API

### 第三步：迁移 Native 代码 (如有)

```
Android JNI/NDK 代码 → cpp/
```

- 将 JNI 代码改为 NAPI
- 在 napi_init.cpp 中注册函数
- 在 types/libentry/index.d.ts 中添加类型声明

### 第四步：创建示例应用

```
Android app/ 示例 → entry/
```

- 在 entry/src/main/ets/pages/ 中创建示例页面
- 展示如何使用库 API
- 在 entry/oh-package.json5 中引用 library

### 第五步：编写测试

```
Android androidTest/ → entry/src/ohosTest/
```

- 将 JUnit 测试改为 hypium 测试
- 覆盖核心功能
- 确保测试通过

### 第六步：配置构建

- 配置 library/build-profile.json5
- 配置 library/oh-package.json5
- 运行 `hvigorw assembleHar` 验证编译

---

## 最佳实践

### 1. 模块化设计

- library/ 只包含核心代码，不包含示例
- entry/ 作为示例应用，展示用法
- 保持 library/ 的 API 简洁明了

### 2. 清晰的导出

- 在 library/Index.ets 中明确导出公共 API
- 不要导出内部实现细节
- 使用 TypeScript 类型提高可读性

### 3. 完善的测试

- 测试覆盖核心功能
- 使用 hypium 断言验证行为
- 定期运行测试确保稳定性

### 4. 文档和示例

- 在 entry/ 中提供丰富的示例代码
- 编写 README.md 说明使用方法
- 提供 API 文档

### 5. 资源管理

- 资源文件放在合适的目录
- 使用 $r() 引用资源
- 支持多语言和主题

---

## 检查清单

迁移完成前，确认以下事项：

- [ ] library/Index.ets 导出所有公共 API
- [ ] library/oh-package.json5 配置正确
- [ ] entry/oh-package.json5 正确引用 library
- [ ] entry/ 中有示例代码展示用法
- [ ] entry/src/ohosTest/ 中有测试用例
- [ ] 如有 Native 代码，cpp/ 配置正确
- [ ] 类型声明文件 (index.d.ts) 完整
- [ ] 资源文件迁移到 resources/ 目录
- [ ] 运行 `hvigorw clean` 和 `hvigorw assembleHar` 成功
- [ ] 运行 `hvigorw --mode module test` 测试通过

---

## 常见问题

### Q1: 如何引用 library 模块？

A: 在 entry/oh-package.json5 中添加依赖：

```json5
{
  "dependencies": {
    "library": "file:../library"
  }
}
```

然后在代码中导入：

```typescript
import { MyClass } from 'library';
```

### Q2: Native 模块如何被调用？

A: 在 ArkTS 中导入 Native 模块：

```typescript
import nativeModule from 'libentry.so';  // 模块名对应 napi_module.nm_modname

let result = nativeModule.add(1, 2);
```

### Q3: 测试代码为什么在 entry/ 而不是 library/?

A: OpenHarmony 的测试框架依赖 Ability 环境，而 library 是 HAR 包，没有 Ability。因此测试代码统一放在 entry/src/ohosTest/。

### Q4: 如何发布 HAR 包？

A: 构建 HAR 包后，可以：

- 本地引用：`"library": "file:../library"`
- 发布到 ohpm：使用 `ohpm publish` 发布到 OpenHarmony 包管理器
- 私有仓库：配置私有 ohpm 仓库地址

### Q5: 资源文件路径如何迁移？

A: Android 的 R.string.xxx 对应 OpenHarmony 的 $r('app.string.xxx')：

```java
// Android
String text = getString(R.string.app_name);
```

```typescript
// OpenHarmony
let text = await this.context.resourceManager.getStringValue($r('app.string.app_name'));
```
