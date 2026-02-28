# OpenHarmony 项目结构与构建系统

本文档覆盖 OpenHarmony 应用/库项目结构、构建配置、ohpm 包管理等内容。移植 Android 库时需按此结构组织代码。

## Table of Contents

- [项目结构对比](#项目结构对比)
- [Module 结构](#module-结构)
- [构建配置](#构建配置)
- [ohpm 包管理](#ohpm-包管理)
- [HAR 与 HSP](#har-与-hsp)
- [Native 项目结构](#native-项目结构)

---

## 项目结构对比

### Android 项目

```
android-library/
├── build.gradle
├── src/
│   ├── main/
│   │   ├── java/com/example/lib/
│   │   ├── res/
│   │   └── AndroidManifest.xml
│   └── test/
└── proguard-rules.pro
```

### OpenHarmony 库项目

```
oh-library/
├── oh-package.json5          # 包描述 (类似 package.json)
├── build-profile.json5       # 构建配置
├── hvigorfile.ts             # 构建脚本
├── src/
│   └── main/
│       ├── ets/              # ArkTS 源码 (替代 java/)
│       │   ├── components/   # UI 组件
│       │   └── utils/        # 工具类
│       ├── resources/        # 资源文件 (替代 res/)
│       │   ├── base/
│       │   │   ├── element/  # 字符串、颜色等
│       │   │   └── media/    # 图片等
│       │   └── rawfile/      # 原始文件
│       └── module.json5      # 模块配置 (替代 AndroidManifest.xml)
├── index.ets                 # 导出入口
└── README.md
```

---

## Module 结构

### module.json5 (核心配置)

```json5
{
  "module": {
    "name": "library",
    "type": "har",                    // har=静态库, hsp=动态共享包, entry=应用入口
    "description": "$string:module_desc",
    "deviceTypes": ["default", "tablet", "2in1"],
    "deliveryWithInstall": true,
    "pages": "$profile:main_pages",   // 仅 entry/feature 模块需要
    "abilities": [],                  // UIAbility 配置
    "extensionAbilities": [],         // ExtensionAbility 配置
    "requestPermissions": [           // 权限声明
      { "name": "ohos.permission.INTERNET" }
    ]
  }
}
```

### Android Manifest 到 module.json5 映射

| AndroidManifest.xml | module.json5 | Notes |
|-------|------------|-------|
| `<uses-permission>` | `requestPermissions` | |
| `<activity>` | `abilities` (type: "page") | |
| `<service>` | `extensionAbilities` | |
| `<receiver>` | 代码注册 commonEvent | |
| `<provider>` | `extensionAbilities` (type: "dataShare") | |
| `minSdkVersion` | `compatibleSdkVersion` in build-profile.json5 | |
| `targetSdkVersion` | `compileSdkVersion` | |

---

## 构建配置

### build-profile.json5 (类似 build.gradle)

```json5
{
  "apiType": "stageModel",
  "buildOption": {
    "arkOptions": {
      "runtimeOnly": {
        "packages": []
      }
    }
  },
  "targets": [
    {
      "name": "default",
      "runtimeOS": "OpenHarmony"
    }
  ]
}
```

### hvigorfile.ts (构建脚本)

```typescript
// HAR 库
import { harTasks } from '@ohos/hvigor-ohos-plugin'
export default { system: harTasks }

// HSP 动态包
import { hspTasks } from '@ohos/hvigor-ohos-plugin'
export default { system: hspTasks }

// Entry 应用
import { hapTasks } from '@ohos/hvigor-ohos-plugin'
export default { system: hapTasks }
```

### Gradle → hvigor 对应关系

| Gradle | hvigor | Notes |
|--------|--------|-------|
| `build.gradle` | `hvigorfile.ts` | 构建脚本 |
| `settings.gradle` | `hvigor/hvigor-config.json5` | 项目级构建配置 |
| `gradle.properties` | `build-profile.json5` | 构建属性 |
| `implementation 'lib:1.0'` | oh-package.json5 `dependencies` | 依赖声明 |
| `api 'lib:1.0'` | oh-package.json5 `dependencies` | OH 不区分 impl/api |
| `apply plugin: 'com.android.library'` | harTasks / hspTasks | 模块类型 |

---

## ohpm 包管理

ohpm (OpenHarmony Package Manager) 类似 npm，用于管理 OH 三方库依赖。

### oh-package.json5 (包描述)

```json5
{
  "name": "@ohos/my-library",
  "version": "1.0.0",
  "description": "移植自 Android 的 XXX 库",
  "main": "index.ets",
  "types": "",
  "author": "",
  "license": "Apache-2.0",
  "dependencies": {
    // 三方依赖
  },
  "devDependencies": {
    // 开发依赖
  }
}
```

### 常用 ohpm 命令

| 命令 | 说明 |
|------|------|
| `ohpm install` | 安装依赖 |
| `ohpm install @ohos/axios` | 安装指定包 |
| `ohpm publish` | 发布包 |
| `ohpm list` | 列出已安装包 |

### 发布到 ohpm

移植完成的库可发布到 ohpm 仓库：

1. 注册 ohpm 账号
2. 配置 oh-package.json5
3. 编写 README.md
4. 执行 `ohpm publish`

---

## HAR 与 HSP

移植 Android 库时，通常选择 HAR 或 HSP 格式：

### HAR (Harmony Archive) — 静态共享包

- 类似 Android 的 AAR
- 编译时打包进使用方
- 适合：工具类库、纯逻辑库、无 UI 资源的库
- 每个使用方有独立副本

### HSP (Harmony Shared Package) — 动态共享包

- 运行时共享，多模块共用一份
- 适合：有大量资源的 UI 组件库
- 减小应用总包体积

### 选择建议

| 场景 | 推荐格式 |
|------|---------|
| 纯逻辑工具库 (如 Gson) | HAR |
| 网络请求封装 | HAR |
| UI 组件库 (如 Material Components) | HSP |
| 包含大量资源/图片 | HSP |
| 仅内部模块间共享 | HSP |
| 发布到 ohpm | HAR |

---

## Native 项目结构

### Android NDK → OpenHarmony Native

```
oh-native-library/
├── oh-package.json5
├── build-profile.json5
├── src/
│   └── main/
│       ├── ets/
│       │   └── utils/
│       │       └── NativeBinding.ets    # ArkTS 绑定层
│       ├── cpp/                          # C/C++ 源码 (替代 jni/)
│       │   ├── CMakeLists.txt
│       │   ├── napi_init.cpp            # NAPI 入口 (替代 JNI_OnLoad)
│       │   ├── types/                   # .d.ts 类型声明
│       │   │   └── libentry/
│       │   │       ├── index.d.ts
│       │   │       └── oh-package.json5
│       │   └── src/                     # 移植的 C/C++ 代码
│       └── module.json5
└── index.ets
```

### CMakeLists.txt 对比

```cmake
# Android NDK CMakeLists.txt
cmake_minimum_required(VERSION 3.18.1)
project("mylib")
add_library(mylib SHARED native-lib.cpp)
find_library(log-lib log)
target_link_libraries(mylib ${log-lib})

# OpenHarmony Native CMakeLists.txt
cmake_minimum_required(VERSION 3.5.0)
project(mylib)
set(NATIVERENDER_ROOT_PATH ${CMAKE_CURRENT_SOURCE_DIR})
add_library(mylib SHARED napi_init.cpp src/mylib.cpp)
target_link_libraries(mylib PUBLIC libace_napi.z.so libhilog_ndk.z.so)
```

---

## hvigorw 构建工具

hvigor 是 OpenHarmony 的构建工具链，使用 hvigorw (hvigor wrapper) 命令行工具编译项目。

### 常用命令

#### 查询命令

| 命令 | 说明 |
|------|------|
| `hvigorw -h, --help` | 打印命令帮助信息 |
| `hvigorw -v, --version, version` | 打印 hvigorw 版本信息 |

#### 编译构建命令

| 命令 | 说明 | 适用场景 |
|------|------|----------|
| `hvigorw clean` | 清理 build 目录 | 完全重新编译前 |
| `hvigorw assembleHar` | 构建 HAR 包 | 构建静态共享库 |
| `hvigorw assembleHsp` | 构建 HSP 包 | 构建动态共享包 |
| `hvigorw assembleHap` | 构建 HAP 应用 | 构建应用模块 |
| `hvigorw assembleApp` | 构建 App 应用 | 构建完整应用 |
| `hvigorw collectCoverage` | 生成覆盖率报表 | 基于打点数据统计 |

### 验证编译

移植完成后，使用 hvigorw 验证编译是否成功：

```bash
# 1. 清理旧的构建产物
hvigorw clean

# 2. 构建 HAR 库
hvigorw assembleHar

# 3. 查看构建产物
# 输出位置: library/build/default/outputs/default/library.har
```

### 编译错误排查

| 错误类型 | 常见原因 | 解决方案 |
|---------|---------|----------|
| ArkTS 语法错误 | Java/Kotlin 语法未完全转换 | 检查类型声明、泛型、可空类型 |
| 模块导入错误 | import 路径错误 | 检查 @ohos.* 模块路径 |
| Native 编译错误 | CMakeLists.txt 配置错误 | 检查库链接、头文件路径 |
| 类型声明错误 | .d.ts 文件缺失或错误 | 为 Native 模块补充类型声明 |
| 循环依赖 | 模块间相互依赖 | 重构模块依赖关系 |

### 构建产物

编译成功后，产物位置：

```
library/build/default/outputs/default/
├── library.har              # HAR 包
└── mapping/                 # 混淆映射 (如启用)
    └── ...
```

### 多模块项目构建

```bash
# 构建指定模块
hvigorw :library:assembleHar

# 构建所有模块
hvigorw assembleApp

# 并行构建 (加速)
hvigorw --parallel assembleApp
```

### Gradle vs hvigorw

| Gradle (Android) | hvigorw (OpenHarmony) | 说明 |
|-----------------|---------------------|------|
| `./gradlew clean` | `hvigorw clean` | 清理 |
| `./gradlew assembleRelease` | `hvigorw assembleHar` | 构建库 |
| `./gradlew build` | `hvigorw assembleApp` | 构建应用 |
| `./gradlew test` | `hvigorw --mode module test` | 运行测试 |
