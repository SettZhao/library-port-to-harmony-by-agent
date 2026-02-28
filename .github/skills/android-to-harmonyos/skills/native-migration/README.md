# Native 代码迁移：JNI/NDK → NAPI

本 skill 覆盖 Android JNI/NDK 代码到 OpenHarmony NAPI (Node-API) 的迁移方法。

---

## 核心迁移步骤

1. 保留 C/C++ 业务逻辑代码不变
2. 删除所有 JNI 绑定代码 (`JNIEnv*`, `JNIEXPORT` 等)
3. 编写 NAPI 入口 (`napi_init.cpp`)，将函数注册到 NAPI
4. **编写 `index.d.ts` 类型声明文件供 ArkTS 调用（关键）**
5. 修改 CMakeLists.txt 链接 OH 系统库

---

## ⚠️ Native 模块类型声明的正确写法（必须遵守）

### CMakeLists.txt
不要修改project和add_library/target_link_libraries中的命名

### index.d.ts 内容（使用具名导出）

**❌ 错误写法（export default）：**
```typescript
export interface ModuleNative {
  funcA: (arg: string) => number;
}
declare const module: ModuleNative;
export default module;
```

**✅ 正确写法（export const）：**
```typescript
/**
 * Function A description
 * @param arg - Parameter description
 * @returns Return value description
 */
export const funcA: (arg: string) => number;
export const funcB: (arg: number) => void;
```

### NAPI 注册代码（必须匹配 index.d.ts）

```cpp
// napi_init.cpp
#include <napi/native_api.h>
#include "hilog/log.h"

static napi_value FuncA(napi_env env, napi_callback_info info) {
    // 实现...
}

EXTERN_C_START
static napi_value Init(napi_env env, napi_value exports) {
    napi_property_descriptor desc[] = {
        // ⚠️ 第一个字符串必须与 index.d.ts 中导出名完全一致
        { "funcA", nullptr, FuncA, nullptr, nullptr, nullptr, napi_default, nullptr },
        { "funcB", nullptr, FuncB, nullptr, nullptr, nullptr, napi_default, nullptr }
    };
    // ⚠️ 必须使用 napi_define_properties
    napi_define_properties(env, exports, sizeof(desc) / sizeof(desc[0]), desc);
    return exports;
}
EXTERN_C_END

static napi_module demoModule = {
    .nm_version = 1,
    .nm_flags = 0,
    .nm_filename = nullptr,
    .nm_register_func = Init,
    .nm_modname = "modulename",
    .nm_priv = ((void*)0),
    .reserved = { 0 },
};

extern "C" __attribute__((constructor)) void RegisterModule(void) {
    napi_module_register(&demoModule);
}
```

### ArkTS 使用方式（使用具名导入）

**❌ 错误：** `import module from 'liblibrary.so';`
**✅ 正确：** `import { funcA, funcB } from 'liblibrary.so';`

```typescript
import { funcA, funcB } from 'liblibrary.so';
const result = funcA('test');  // ✅ 正常工作
```

## 处理 Native 库的外部依赖

### 检测依赖

```bash
grep -r "#include <" <library>/src/ | grep -v "std\|stdio\|stdlib\|string"
```

### 处理策略

1. **简单工具函数/数据结构** → 实现简化版本
2. **复杂三方库** → 一并迁移或寻找 OH 替代
3. **未使用的依赖** → `#if 0 ... #endif` 注释

---

## hilog 日志格式化符

**⚠️ 必须使用 `%{public}` 修饰符，否则日志参数显示为 `<private>`**

```c
// ❌ 错误：参数不可见
OH_LOG_INFO(LOG_APP, "Found %u entries", count);

// ✅ 正确：参数可见
OH_LOG_INFO(LOG_APP, "Found %{public}u entries", count);
OH_LOG_DEBUG(LOG_APP, "Value: %{public}s = %{public}d", key, val);
```

### 处理日志宏冲突

```c
#include <hilog/log.h>
#undef LOG_DOMAIN
#undef LOG_TAG
#define LOG_DOMAIN 0x0001
#define LOG_TAG "libmp4"

#define MP4_LOGD(...) OH_LOG_DEBUG(LOG_APP, __VA_ARGS__)
#define MP4_LOGI(...) OH_LOG_INFO(LOG_APP, __VA_ARGS__)
#define MP4_LOGE(...) OH_LOG_ERROR(LOG_APP, __VA_ARGS__)
```

---

## 数据类型转换速查

| JNI 类型 | NAPI 类型 | 创建函数 | 获取函数 |
|----------|----------|---------|---------|
| `jint` | `napi_value` (number) | `napi_create_int32()` | `napi_get_value_int32()` |
| `jlong` | `napi_value` (number) | `napi_create_int64()` | `napi_get_value_int64()` |
| `jdouble` | `napi_value` (number) | `napi_create_double()` | `napi_get_value_double()` |
| `jboolean` | `napi_value` (boolean) | `napi_get_boolean()` | `napi_get_value_bool()` |
| `jstring` | `napi_value` (string) | `napi_create_string_utf8()` | `napi_get_value_string_utf8()` |
| `jbyteArray` | `napi_value` (ArrayBuffer) | `napi_create_arraybuffer()` | `napi_get_arraybuffer_info()` |
| `jobject` | `napi_value` (object) | `napi_create_object()` | `napi_get_named_property()` |

---

## 验证清单

- [ ] 所有函数使用 `export const funcName: ...` 声明
- [ ] 没有使用 `export default` 或 `export interface`
- [ ] ArkTS 代码使用具名导入 `import { func } from 'lib.so'`

---

## 详细参考

- [references/native-migration.md](../../references/native-migration.md) — JNI→NAPI 完整迁移示例
