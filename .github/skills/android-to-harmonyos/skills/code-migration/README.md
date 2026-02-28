# 代码迁移：Java/Kotlin → ArkTS

本 skill 覆盖移植工作流的 Step 4 中通用代码迁移部分：将 Java/Kotlin 代码翻译为 ArkTS。

---

## 通用迁移规则

1. Java/Kotlin class → ArkTS class 或 struct (UI 组件用 `@Component` struct)
2. Android import → OpenHarmony import
3. `Log.d()` → `hilog.debug()`
4. `SharedPreferences` → `@ohos.data.preferences`
5. `Intent` → `Want`
6. `Activity/Fragment` lifecycle → `UIAbility` lifecycle

---

## 代码翻译要点

### 类型系统

| Java/Kotlin | ArkTS | 说明 |
|------------|-------|------|
| `int` / `long` | `number` | 所有数值统一为 number |
| `float` / `double` | `number` | |
| `boolean` | `boolean` | |
| `String` | `string` | 小写 |
| `List<T>` | `Array<T>` | |
| `Map<K,V>` | `Map<K,V>` | ArkTS 内置 Map |
| `Set<T>` | `Set<T>` | ArkTS 内置 Set |
| `byte[]` | `ArrayBuffer` / `Uint8Array` | 二进制数据 |
| `void` | `void` | |
| `Object` | `object` / `Record<string, Object>` | |

### 空安全

```java
// Java
@Nullable String getName() { ... }
if (name != null) { use(name); }

// ArkTS
getName(): string | null { ... }
if (name !== null) { use(name); }
```

```kotlin
// Kotlin
val name: String? = obj?.name
name?.let { use(it) }

// ArkTS
let name: string | undefined = obj?.name;
if (name) { use(name); }
```

### 泛型

```java
// Java
public class Cache<K, V> {
    private Map<K, V> map = new HashMap<>();
    public V get(K key) { return map.get(key); }
}

// ArkTS
class Cache<K, V> {
  private map: Map<K, V> = new Map();
  get(key: K): V | undefined { return this.map.get(key); }
}
```

### 接口

```java
// Java (interface with default method)
interface Callback {
    void onSuccess(String data);
    default void onError(Exception e) { log(e); }
}

// ArkTS (无 default method，改用抽象类或工具函数)
interface Callback {
  onSuccess(data: string): void;
  onError(e: Error): void;
}

// 或提供默认实现的抽象类
abstract class BaseCallback implements Callback {
  onError(e: Error): void { /* default impl */ }
}
```

### 异常处理

```java
// Java
try {
    doSomething();
} catch (IOException e) {
    handleError(e);
} finally {
    cleanup();
}

// ArkTS (完全一致)
try {
  doSomething();
} catch (e) {
  handleError(e as Error);
} finally {
  cleanup();
}
```

### 异步处理

```java
// Android (Coroutines)
lifecycleScope.launch {
    val result = withContext(Dispatchers.IO) { heavyWork() }
    updateUI(result)
}

// ArkTS (async/await + TaskPool)
import { taskpool } from '@kit.ArkTS';

@Concurrent
function heavyWork(): string { return 'result'; }

let task = new taskpool.Task(heavyWork);
let result = await taskpool.execute(task) as string;
// UI update happens on main thread after await
```

### 常用模式转换

```java
// Java - Singleton
public class AppConfig {
    private static AppConfig instance;
    public static synchronized AppConfig getInstance() {
        if (instance == null) instance = new AppConfig();
        return instance;
    }
}

// ArkTS - Singleton
class AppConfig {
  private static instance: AppConfig | null = null;
  static getInstance(): AppConfig {
    if (!AppConfig.instance) { AppConfig.instance = new AppConfig(); }
    return AppConfig.instance;
  }
}
```

```java
// Java - Builder Pattern
MyObj obj = new MyObj.Builder()
    .setName("test")
    .setValue(42)
    .build();

// ArkTS - Object literal or Builder
let obj: MyObj = { name: "test", value: 42 };
// 或保留 Builder 模式
```

---

## 常用 API 替换速查

| Android | OpenHarmony | import |
|---------|-------------|--------|
| `Log.d(TAG, msg)` | `hilog.debug(0x0000, TAG, msg)` | `import { hilog } from '@kit.PerformanceAnalysisKit'` |
| `SharedPreferences` | `preferences` | `import dataPreferences from '@ohos.data.preferences'` |
| `HttpURLConnection` | `http.createHttp()` | `import http from '@ohos.net.http'` |
| `Thread` / `Executor` | `TaskPool` | `import { taskpool } from '@kit.ArkTS'` |
| `Intent` | `Want` | `import Want from '@ohos.app.ability.Want'` |
| `Toast` | `promptAction.showToast()` | `import promptAction from '@ohos.promptAction'` |
| `JSON (Gson)` | `JSON.parse()` / `JSON.stringify()` | 内置 |

---