# Native Migration: JNI/NDK → NAPI

本文档覆盖 Android JNI/NDK 代码到 OpenHarmony NAPI (Node-API) 的迁移方法。

## Table of Contents

- [概述对比](#概述对比)
- [JNI → NAPI 核心概念](#jni--napi-核心概念)
- [函数注册](#函数注册)
- [数据类型转换](#数据类型转换)
- [回调与异步](#回调与异步)
- [CMake 构建迁移](#cmake-构建迁移)
- [常见 NDK API 映射](#常见-ndk-api-映射)
- [完整迁移示例](#完整迁移示例)

---

## 概述对比

| 维度 | Android JNI | OpenHarmony NAPI |
|------|------------|-----------------|
| 接口标准 | JNI (Java Native Interface) | Node-API (NAPI) |
| 绑定语言 | Java/Kotlin ↔ C/C++ | ArkTS/JS ↔ C/C++ |
| 注册方式 | `JNI_OnLoad` / `RegisterNatives` | `napi_module_register` |
| 函数命名 | `Java_com_pkg_Class_method` | 自定义，模块注册时绑定 |
| 数据传递 | `jstring`, `jint`, `jobject` 等 | `napi_value` 统一类型 |
| 线程模型 | `AttachCurrentThread` | `napi_create_async_work` |
| 头文件 | `<jni.h>` | `<napi/native_api.h>` |

---

## JNI → NAPI 核心概念

### 入口注册

```c
// Android JNI
JNIEXPORT jint JNI_OnLoad(JavaVM *vm, void *reserved) {
    JNIEnv *env;
    vm->GetEnv((void**)&env, JNI_VERSION_1_6);
    // 注册 native 方法
    return JNI_VERSION_1_6;
}

// 或通过命名约定
JNIEXPORT jstring JNICALL
Java_com_example_MyClass_nativeMethod(JNIEnv *env, jobject thiz) {
    return env->NewStringUTF("Hello from JNI");
}
```

```c
// OpenHarmony NAPI
#include <napi/native_api.h>

static napi_value NativeMethod(napi_env env, napi_callback_info info) {
    napi_value result;
    napi_create_string_utf8(env, "Hello from NAPI", NAPI_AUTO_LENGTH, &result);
    return result;
}

// 模块注册
static napi_value Init(napi_env env, napi_value exports) {
    napi_property_descriptor desc[] = {
        { "nativeMethod", nullptr, NativeMethod, nullptr, nullptr, nullptr,
          napi_default, nullptr }
    };
    napi_define_properties(env, exports, sizeof(desc) / sizeof(desc[0]), desc);
    return exports;
}

static napi_module myModule = {
    .nm_version = 1,
    .nm_flags = 0,
    .nm_filename = nullptr,
    .nm_register_func = Init,
    .nm_modname = "mylib",
    .nm_priv = nullptr,
    .reserved = { 0 },
};

extern "C" __attribute__((constructor)) void RegisterModule(void) {
    napi_module_register(&myModule);
}
```

---

## 函数注册

### JNI 动态注册 → NAPI 注册

```c
// JNI 动态注册
static JNINativeMethod methods[] = {
    {"add",        "(II)I",                  (void*)native_add},
    {"getMessage", "()Ljava/lang/String;",   (void*)native_getMessage},
    {"process",    "([B)[B",                 (void*)native_process},
};

// NAPI 注册
static napi_value Init(napi_env env, napi_value exports) {
    napi_property_descriptor desc[] = {
        { "add", nullptr, NativeAdd, nullptr, nullptr, nullptr, napi_default, nullptr },
        { "getMessage", nullptr, NativeGetMessage, nullptr, nullptr, nullptr, napi_default, nullptr },
        { "process", nullptr, NativeProcess, nullptr, nullptr, nullptr, napi_default, nullptr },
    };
    napi_define_properties(env, exports, sizeof(desc) / sizeof(desc[0]), desc);
    return exports;
}
```

---

## 数据类型转换

### JNI 类型 → NAPI 类型

| JNI | NAPI | 创建函数 | 获取函数 |
|-----|------|---------|---------|
| `jint` | `napi_value` (number) | `napi_create_int32()` | `napi_get_value_int32()` |
| `jlong` | `napi_value` (bigint/number) | `napi_create_int64()` | `napi_get_value_int64()` |
| `jdouble` | `napi_value` (number) | `napi_create_double()` | `napi_get_value_double()` |
| `jboolean` | `napi_value` (boolean) | `napi_get_boolean()` | `napi_get_value_bool()` |
| `jstring` | `napi_value` (string) | `napi_create_string_utf8()` | `napi_get_value_string_utf8()` |
| `jbyteArray` | `napi_value` (ArrayBuffer) | `napi_create_arraybuffer()` | `napi_get_arraybuffer_info()` |
| `jobject` | `napi_value` (object) | `napi_create_object()` | `napi_get_named_property()` |
| `jarray` | `napi_value` (Array) | `napi_create_array()` | `napi_get_element()` |

### 参数获取

```c
// JNI
JNIEXPORT jint JNICALL
Java_com_example_Lib_add(JNIEnv *env, jobject thiz, jint a, jint b) {
    return a + b;
}

// NAPI
static napi_value NativeAdd(napi_env env, napi_callback_info info) {
    size_t argc = 2;
    napi_value args[2];
    napi_get_cb_info(env, info, &argc, args, nullptr, nullptr);

    int32_t a, b;
    napi_get_value_int32(env, args[0], &a);
    napi_get_value_int32(env, args[1], &b);

    napi_value result;
    napi_create_int32(env, a + b, &result);
    return result;
}
```

### String 处理

```c
// JNI
const char *str = env->GetStringUTFChars(jstr, nullptr);
// use str...
env->ReleaseStringUTFChars(jstr, str);

// NAPI
size_t len;
napi_get_value_string_utf8(env, args[0], nullptr, 0, &len);
char *str = new char[len + 1];
napi_get_value_string_utf8(env, args[0], str, len + 1, &len);
// use str...
delete[] str;
```

### byte[] / ArrayBuffer 处理

```c
// JNI
jbyte *bytes = env->GetByteArrayElements(jByteArray, nullptr);
jsize length = env->GetArrayLength(jByteArray);
// use bytes...
env->ReleaseByteArrayElements(jByteArray, bytes, 0);

// NAPI (ArrayBuffer)
void *data;
size_t length;
napi_get_arraybuffer_info(env, args[0], &data, &length);
uint8_t *bytes = (uint8_t *)data;
// use bytes...
```

---

## 回调与异步

### JNI 异步回调 → NAPI 异步

```c
// JNI - 子线程回调 Java
void *thread_func(void *arg) {
    JNIEnv *env;
    g_jvm->AttachCurrentThread(&env, nullptr);
    // call Java callback
    env->CallVoidMethod(g_callback, g_method, result);
    g_jvm->DetachCurrentThread();
    return nullptr;
}

// NAPI - 异步任务
typedef struct {
    napi_async_work work;
    napi_ref callbackRef;
    int result;
} AsyncData;

void ExecuteWork(napi_env env, void *data) {
    AsyncData *d = (AsyncData *)data;
    d->result = heavyComputation();  // 在子线程执行
}

void CompleteWork(napi_env env, napi_status status, void *data) {
    AsyncData *d = (AsyncData *)data;
    napi_value callback, result, retval;
    napi_get_reference_value(env, d->callbackRef, &callback);
    napi_create_int32(env, d->result, &result);
    napi_call_function(env, nullptr, callback, 1, &result, &retval);
    napi_delete_async_work(env, d->work);
    napi_delete_reference(env, d->callbackRef);
    free(d);
}

static napi_value StartAsyncWork(napi_env env, napi_callback_info info) {
    size_t argc = 1;
    napi_value args[1];
    napi_get_cb_info(env, info, &argc, args, nullptr, nullptr);

    AsyncData *data = (AsyncData *)malloc(sizeof(AsyncData));
    napi_create_reference(env, args[0], 1, &data->callbackRef);

    napi_value workName;
    napi_create_string_utf8(env, "asyncWork", NAPI_AUTO_LENGTH, &workName);
    napi_create_async_work(env, nullptr, workName, ExecuteWork, CompleteWork, data, &data->work);
    napi_queue_async_work(env, data->work);

    return nullptr;
}
```

### ArkTS 调用 NAPI

```typescript
// index.d.ts (类型声明)
export const add: (a: number, b: number) => number;
export const getMessage: () => string;
export const startAsyncWork: (callback: (result: number) => void) => void;

// ArkTS 使用
import mylib from 'libmylib.so'

let sum = mylib.add(1, 2)
let msg = mylib.getMessage()
mylib.startAsyncWork((result: number) => {
  console.log('Async result: ' + result)
})
```

---

## CMake 构建迁移

### 关键差异

| Android NDK | OpenHarmony NDK | Notes |
|-------------|----------------|-------|
| `${ANDROID_ABI}` | `${OHOS_ARCH}` | arm64-v8a → arm64 |
| `-landroid` | 无 | |
| `-llog` | `-lhilog_ndk.z` | 日志库 |
| `-ljnigraphics` | `-lnative_image` / `-lpixelmap_ndk` | 图像处理 |
| `-lOpenSLES` | `-lOHAudio` | 音频 |
| `-lEGL -lGLESv3` | `-lEGL -lGLESv3` | 相同 |
| `-lvulkan` | `-lvulkan` | 相同 |
| `ANDROID_NDK` | `OHOS_SDK` | SDK path |

### CMakeLists.txt 迁移模板

```cmake
cmake_minimum_required(VERSION 3.5.0)
project(mylib)

# 源文件 (保留原有 C/C++ 逻辑代码)
set(SRC_FILES
    napi_init.cpp          # 新增: NAPI 入口
    src/original_code.cpp  # 原有: 业务逻辑
)

add_library(mylib SHARED ${SRC_FILES})

# OpenHarmony SDK 系统库
target_link_libraries(mylib PUBLIC
    libace_napi.z.so       # NAPI 核心
    libhilog_ndk.z.so      # 日志 (替代 Android log)
    # 按需添加:
    # libnative_image.so   # 图像处理
    # libOHAudio.so        # 音频
    # libnative_drawing.so # 2D 绘制
)

# 保留原有头文件路径
target_include_directories(mylib PUBLIC
    ${CMAKE_CURRENT_SOURCE_DIR}/src
    ${CMAKE_CURRENT_SOURCE_DIR}/include
)
```

---

## 常见 NDK API 映射

| Android NDK | OpenHarmony NDK | 头文件 |
|-------------|----------------|--------|
| `__android_log_print` | `OH_LOG_Print` | `<hilog/log.h>` |
| `ANativeWindow` | `OHNativeWindow` | `<native_window/external_window.h>` |
| `AImageReader` | `OH_ImageReceiverNative` | `<multimedia/image_framework/>` |
| `ASensor*` | `OH_Sensor*` | `<sensors/oh_sensor.h>` |
| `AAudio*` | `OH_Audio*` | `<ohaudio/native_audiostreambuilder.h>` |
| `AMediaCodec` | `OH_AVCodec` | `<multimedia/player_framework/>` |
| `AHardwareBuffer` | `OH_NativeBuffer` | `<native_buffer/native_buffer.h>` |

---

## 完整迁移示例

### Android JNI 库 → OpenHarmony NAPI 库

**原始 Android 代码:**

```java
// Java
public class CryptoLib {
    static { System.loadLibrary("crypto"); }
    public native byte[] encrypt(byte[] data, String key);
    public native byte[] decrypt(byte[] data, String key);
}
```

```c
// JNI C
JNIEXPORT jbyteArray JNICALL
Java_com_example_CryptoLib_encrypt(JNIEnv *env, jobject thiz,
    jbyteArray data, jstring key) {
    jbyte *dataBytes = (*env)->GetByteArrayElements(env, data, NULL);
    jsize dataLen = (*env)->GetArrayLength(env, data);
    const char *keyStr = (*env)->GetStringUTFChars(env, key, NULL);

    // 核心加密逻辑 (可直接复用)
    uint8_t *result = do_encrypt((uint8_t*)dataBytes, dataLen, keyStr);
    size_t resultLen = get_result_len();

    jbyteArray output = (*env)->NewByteArray(env, resultLen);
    (*env)->SetByteArrayRegion(env, output, 0, resultLen, (jbyte*)result);

    (*env)->ReleaseByteArrayElements(env, data, dataBytes, 0);
    (*env)->ReleaseStringUTFChars(env, key, keyStr);
    free(result);
    return output;
}
```

**迁移后 OpenHarmony 代码:**

```c
// napi_init.cpp
#include <napi/native_api.h>
#include "crypto_core.h"  // 复用原有加密逻辑

static napi_value Encrypt(napi_env env, napi_callback_info info) {
    size_t argc = 2;
    napi_value args[2];
    napi_get_cb_info(env, info, &argc, args, nullptr, nullptr);

    // 获取 ArrayBuffer 数据
    void *data;
    size_t dataLen;
    napi_get_arraybuffer_info(env, args[0], &data, &dataLen);

    // 获取 key 字符串
    size_t keyLen;
    napi_get_value_string_utf8(env, args[1], nullptr, 0, &keyLen);
    char *key = new char[keyLen + 1];
    napi_get_value_string_utf8(env, args[1], key, keyLen + 1, &keyLen);

    // 复用原有加密逻辑
    uint8_t *result = do_encrypt((uint8_t*)data, dataLen, key);
    size_t resultLen = get_result_len();

    // 返回 ArrayBuffer
    void *resultData;
    napi_value output;
    napi_create_arraybuffer(env, resultLen, &resultData, &output);
    memcpy(resultData, result, resultLen);

    delete[] key;
    free(result);
    return output;
}

static napi_value Init(napi_env env, napi_value exports) {
    napi_property_descriptor desc[] = {
        { "encrypt", nullptr, Encrypt, nullptr, nullptr, nullptr, napi_default, nullptr },
        { "decrypt", nullptr, Decrypt, nullptr, nullptr, nullptr, napi_default, nullptr },
    };
    napi_define_properties(env, exports, sizeof(desc) / sizeof(desc[0]), desc);
    return exports;
}

EXTERN_C_START
static napi_module cryptoModule = {
    .nm_version = 1, .nm_flags = 0, .nm_filename = nullptr,
    .nm_register_func = Init, .nm_modname = "crypto", .nm_priv = nullptr,
};
__attribute__((constructor)) void RegisterModule() { napi_module_register(&cryptoModule); }
EXTERN_C_END
```

```typescript
// index.d.ts
export const encrypt: (data: ArrayBuffer, key: string) => ArrayBuffer;
export const decrypt: (data: ArrayBuffer, key: string) => ArrayBuffer;

// ArkTS 使用
import crypto from 'libcrypto.so'
let encrypted = crypto.encrypt(dataBuffer, 'myKey')
```
