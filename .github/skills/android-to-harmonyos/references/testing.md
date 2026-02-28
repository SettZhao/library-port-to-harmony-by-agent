# OpenHarmony 测试框架

## @ohos/hypium 测试框架

hypium 是 OpenHarmony 的单元测试和 UI 测试框架，类似于 Android 的 JUnit + Espresso。

### 依赖配置

在 oh-package.json5 中添加测试依赖：

```json
{
  "devDependencies": {
    "@ohos/hypium": "^1.0.4"
  }
}
```

### 测试文件结构

```
entry/src/ohosTest/ets/
├── test/
│   ├── Ability.test.ets      # 测试入口
│   └── List.test.ets          # 测试套件列表
└── testability/
    └── TestAbility.ets        # 测试 Ability
```

### 基本测试示例

```typescript
import { describe, it, expect } from '@ohos/hypium';

export default function abilityTest() {
  describe('MyLibrary', () => {
    
    it('shouldAddTwoNumbersCorrectly', () => {
      let result = add(2, 3);
      expect(result).assertEqual(5);
    });
    
    it('shouldHandleNullInput', () => {
      let result = processInput(null);
      expect(result).assertNull();
    });
    
    it('shouldReturnValidObject', () => {
      let obj = createObject();
      expect(obj).not().assertNull(); // 使用 not() 实现 assertNotNull
    });
    
  });
}
```

**⚠️ 重要：测试命名规范**

OpenHarmony 测试框架对命名有严格要求：

1. **`describe()` 测试套件名称规范：**
   - ✅ **必须符合**：只能包含数字、字母、下划线 `_` 和点 `.`
   - ✅ **必须以字母开头**
   - ❌ **不能包含空格** 或其他特殊字符
   - 推荐使用**驼峰命名**（CamelCase）
   
   **示例：**
   - ❌ 错误：`describe('My Library Tests', ...)` — 包含空格
   - ❌ 错误：`describe('My-Library', ...)` — 包含连字符
   - ❌ 错误：`describe('123Test', ...)` — 以数字开头
   - ✅ 正确：`describe('MyLibraryTests', ...)` — 驼峰命名
   - ✅ 正确：`describe('my_library_tests', ...)` — 下划线分隔
   - ✅ 正确：`describe('library.core.tests', ...)` — 点分隔

2. **`it()` 测试用例名称规范：**
   - ✅ **必须符合**：不能包含空格
   - 推荐使用**驼峰命名**（camelCase）
   
   **示例：**
   - ❌ 错误：`it('should add numbers', ...)` — 包含空格
   - ✅ 正确：`it('shouldAddNumbers', ...)` — 驼峰命名
   - ✅ 正确：`it('testUserLogin', ...)` — 驼峰命名
   - ✅ 正确：`it('test001', ...)` — 无分隔符

**命名规范总结：**
- `describe()` 和 `it()` 的名称都不能包含空格
- `describe()` 只能包含字母、数字、下划线、点，且必须以字母开头
- `it()` 推荐使用驼峰命名
- 用例名应清晰描述测试的预期行为

---

## expect 断言方法

hypium 提供了丰富的断言方法，用于验证测试结果。

### 基本断言

| 方法 | 说明 | 示例 |
|------|------|------|
| `assertEqual(value)` | 断言相等 | `expect(result).assertEqual(5)` |
| `assertTrue()` | 断言为 true | `expect(flag).assertTrue()` |
| `assertFalse()` | 断言为 false | `expect(flag).assertFalse()` |
| `assertFail(message)` | 抛出错误，使测试失败 | `expect().assertFail('Unexpected error')` |

**⚠️ 断言不相等：** 使用 `not().assertEqual(value)`，例如：`expect(result).not().assertEqual(0)`

### 空值断言

| 方法 | 说明 | 示例 |
|------|------|------|
| `assertNull()` | 断言为 null | `expect(obj).assertNull()` |
| `assertUndefined()` | 断言为 undefined | `expect(value).assertUndefined()` |
| `assertNaN()` | 断言为 NaN | `expect(result).assertNaN()` |
| `assertNegUnlimited()` | 断言为负无穷 | `expect(value).assertNegUnlimited()` |
| `assertPosUnlimited()` | 断言为正无穷 | `expect(value).assertPosUnlimited()` |

**⚠️ 取反断言（not）**

从 @ohos/hypium 1.0.4 开始支持 `not()` 方法，用于对所有断言取反：

| 需求 | 写法 | 示例 |
|------|------|------|
| 断言不为 null | `not().assertNull()` | `expect(obj).not().assertNull()` |
| 断言不为 undefined | `not().assertUndefined()` | `expect(value).not().assertUndefined()` |
| 断言不为 NaN | `not().assertNaN()` | `expect(result).not().assertNaN()` |
| 断言不相等 | `not().assertEqual(value)` | `expect(result).not().assertEqual(0)` |
| 断言不包含 | `not().assertContain(element)` | `expect(list).not().assertContain('item')` |

### 类型断言

| 方法 | 说明 | 示例 |
|------|------|------|
| `assertInstanceOf(className)` | 断言为指定类型实例 | `expect(obj).assertInstanceOf('MyClass')` |

### 数值断言

| 方法 | 说明 | 示例 |
|------|------|------|
| `assertLarger(value)` | 断言大于指定值 | `expect(score).assertLarger(60)` |
| `assertLargerOrEqual(value)` | 断言大于或等于指定值 | `expect(score).assertLargerOrEqual(60)` |
| `assertLess(value)` | 断言小于指定值 | `expect(age).assertLess(100)` |
| `assertLessOrEqual(value)` | 断言小于或等于指定值 | `expect(age).assertLessOrEqual(100)` |
| `assertClose(value, delta)` | 断言接近指定值（误差范围内） | `expect(pi).assertClose(3.14, 0.01)` |

### 集合断言

| 方法 | 说明 | 示例 |
|------|------|------|
| `assertContain(element)` | 断言包含指定元素 | `expect(list).assertContain('item')` |
| `assertDeepEquals(object)` | 断言深度相等（递归比较） | `expect(obj1).assertDeepEquals(obj2)` |

**⚠️ 断言不包含：** 使用 `not().assertContain(element)`，例如：`expect(list).not().assertContain('item')`

### 异常断言

| 方法 | 说明 | 示例 |
|------|------|------|
| `assertThrowError(errorMsg)` | 断言抛出错误 | `expect(() => divide(1, 0)).assertThrowError('Division by zero')` |

### Promise 断言

| 方法 | 说明 | 示例 |
|------|------|------|
| `assertPromiseIsPending()` | 断言 Promise 为 pending 状态 | `expect(promise).assertPromiseIsPending()` |
| `assertPromiseIsRejected()` | 断言 Promise 为 rejected 状态 | `expect(promise).assertPromiseIsRejected()` |
| `assertPromiseIsRejectedWith(value)` | 断言 Promise 被拒绝并返回指定值 | `expect(promise).assertPromiseIsRejectedWith('error')` |
| `assertPromiseIsRejectedWithError(value)` | 断言 Promise 被拒绝并抛出错误 | `expect(promise).assertPromiseIsRejectedWithError('error')` |
| `assertPromiseIsResolved()` | 断言 Promise 为 resolved 状态 | `expect(promise).assertPromiseIsResolved()` |
| `assertPromiseIsResolvedWith(value)` | 断言 Promise 成功并返回指定值 | `expect(promise).assertPromiseIsResolvedWith(result)` |

---

## Android vs OpenHarmony 测试对比

### 框架对比

| Android | OpenHarmony | 说明 |
|---------|-------------|------|
| JUnit | @ohos/hypium | 单元测试框架 |
| `@Test` | `it()` | 测试方法 |
| `@Before` | `beforeAll()`, `beforeEach()` | 前置操作 |
| `@After` | `afterAll()`, `afterEach()` | 后置操作 |
| `assertEquals()` | `expect().assertEqual()` | 断言相等 |
| `assertNotEquals()` | `expect().not().assertEqual()` | 断言不相等 |
| `assertTrue()` | `expect().assertTrue()` | 断言真值 |
| `assertNull()` | `expect().assertNull()` | 断言空值 |
| `assertNotNull()` | `expect().not().assertNull()` | 断言非空值 |

### 断言语法对比

```java
// Android (JUnit)
@Test
public void testAdd() {
    int result = Calculator.add(2, 3);
    assertEquals(5, result);
    assertTrue(result > 0);
    assertNotNull(result);
}
```

```typescript
// OpenHarmony (hypium)
it('testAdd', () => {
  let result = Calculator.add(2, 3);
  expect(result).assertEqual(5);
  expect(result > 0).assertTrue();
  expect(result).not().assertNull(); // 使用 not() 实现 assertNotNull
});
```

**⚠️ 命名规范：** OpenHarmony 测试用例名称不能包含空格，必须使用驼峰命名（camelCase）或无分隔符命名。

---

## 生命周期钩子

hypium 提供测试生命周期钩子函数：

| 钩子函数 | 执行时机 | 说明 |
|---------|---------|------|
| `beforeAll()` | 所有测试前执行一次 | 初始化资源 |
| `beforeEach()` | 每个测试前执行 | 准备测试环境 |
| `afterEach()` | 每个测试后执行 | 清理测试环境 |
| `afterAll()` | 所有测试后执行一次 | 释放资源 |

### 示例

```typescript
import { describe, it, expect, beforeAll, beforeEach, afterEach, afterAll } from '@ohos/hypium';

export default function testSuite() {
  describe('Database Tests', () => {
    
    beforeAll(() => {
      console.info('Setup database connection');
      // 初始化数据库连接
    });
    
    beforeEach(() => {
      console.info('Clear test data');
      // 每个测试前清空数据
    });
    
    it('shouldInsertRecord', () => {
      let db = getDatabase();
      db.insert({id: 1, name: 'test'});
      expect(db.count()).assertEqual(1);
    });
    
    it('shouldDeleteRecord', () => {
      let db = getDatabase();
      db.insert({id: 1, name: 'test'});
      db.delete(1);
      expect(db.count()).assertEqual(0);
    });
    
    afterEach(() => {
      console.info('Rollback transaction');
      // 回滚数据库事务
    });
    
    afterAll(() => {
      console.info('Close database connection');
      // 关闭数据库连接
    });
    
  });
}
```

---

## 异步测试

### Promise 测试

```typescript
it('shouldFetchDataFromAPI', async () => {
  let promise = fetchData();
  
  // 等待 Promise 完成
  await promise;
  
  // 断言 Promise 状态
  expect(promise).assertPromiseIsResolved();
  expect(promise).assertPromiseIsResolvedWith({status: 'ok'});
});
```

### 超时控制

```typescript
it('shouldTimeoutAfter5Seconds', () => {
  // 设置测试超时时间（毫秒）
  jest.setTimeout(5000);
  
  let promise = longRunningTask();
  expect(promise).assertPromiseIsRejected();
}, 5000); // 第二个参数指定超时
```

---

## UI 测试

hypium 也支持 UI 组件测试：

```typescript
import { Driver, ON } from '@ohos.UiTest';

it('shouldDisplayWelcomeText', async () => {
  // 创建 UI 驱动
  let driver = Driver.create();
  
  // 查找文本组件
  let text = await driver.findComponent(ON.text('Welcome'));
  
  // 断言组件存在
  expect(text).not().assertNull();
  
  // 点击按钮
  let button = await driver.findComponent(ON.id('btn_start'));
  await button.click();
  
  // 验证页面跳转
  let newText = await driver.findComponent(ON.text('Started'));
  expect(newText).not().assertNull();
});
```

---

## 测试覆盖率

### 生成覆盖率报告

```bash
# 构建并运行测试
hvigorw --mode module test

# 收集覆盖率数据
hvigorw collectCoverage
```

### 查看覆盖率报告

覆盖率报告位置：

```
entry/build/default/outputs/coverage/
├── html/                    # HTML 格式报告
│   └── index.html
└── lcov.info               # LCOV 格式数据
```

---

## 运行测试

### DevEco Studio 运行

1. 右键点击测试文件
2. 选择 "Run 'Ability.test.ets'"
3. 查看测试结果面板

### 命令行运行

```bash
# 运行所有测试
hvigorw --mode module test

# 运行指定测试套件
hvigorw --mode module test --test-file Ability.test.ets
```

---

## 测试最佳实践

### 1. 命名规范

- **测试文件**：`*.test.ets`
- **测试套件**：`describe('TestSuiteName', () => {})`
  - ⚠️ **只能包含**：字母、数字、下划线 `_`、点 `.`
  - ⚠️ **必须以字母开头**
  - ⚠️ **不能包含空格**或其他特殊字符
  - 推荐使用**驼峰命名**（CamelCase）
  - ❌ 错误：`describe('My Library Tests', ...)` —— 包含空格
  - ✅ 正确：`describe('MyLibraryTests', ...)` —— 驼峰命名
- **测试用例**：`it('testCaseName', () => {})`
  - ⚠️ **不能包含空格**
  - 推荐使用**驼峰命名**（camelCase）
  - ❌ 错误：`it('should add numbers', ...)` —— 包含空格
  - ✅ 正确：`it('shouldAddNumbers', ...)` —— 驼峰命名
- **描述清晰**：用例名应清晰描述测试的预期行为

### 2. 独立性

- 每个测试用例应独立运行
- 不依赖其他测试的执行顺序
- 使用 `beforeEach()` 准备初始状态

### 3. 断言准确

- 一个测试用例关注一个行为
- 使用最准确的断言方法
- 避免过度测试实现细节

### 4. 异步处理

- 使用 `async/await` 处理异步操作
- 为异步测试设置合理超时
- 确保 Promise 正确处理

### 5. Mock 和 Stub

```typescript
// Mock 网络请求
import { http } from '@ohos.net.http';

// 模拟 HTTP 响应
jest.spyOn(http, 'createHttp').mockReturnValue({
  request: () => Promise.resolve({
    result: 'success',
    responseCode: 200
  })
});
```

---

## 迁移检查清单

从 Android 测试迁移到 OpenHarmony 时：

- [ ] 将 `@Test` 注解改为 `it()` 函数
- [ ] **⚠️ 将测试套件名改为符合规范的命名**：`describe('TestSuiteName', ...)`
  - 只能包含字母、数字、下划线、点，必须以字母开头
  - 不能包含空格（如 `'My Tests'` → `'MyTests'`）
- [ ] **⚠️ 将测试用例名改为驼峰命名（不能有空格）**：`it('shouldAddNumbers', ...)`
- [ ] 将 `@Before/@After` 改为 `beforeEach()/afterEach()`
- [ ] 将 `assertEquals(expected, actual)` 改为 `expect(actual).assertEqual(expected)`
- [ ] 将 `assertTrue(condition)` 改为 `expect(condition).assertTrue()`
- [ ] 将 `assertNotNull(value)` 改为 `expect(value).not().assertNull()`（使用 `not()` 取反）
- [ ] 将 Mockito mock 改为 hypium 的 mock 方式
- [ ] 异步测试使用 `async/await` 替代回调
- [ ] UI 测试使用 `@ohos.UiTest` 替代 Espresso
- [ ] 添加 `@ohos/hypium` 依赖到 `oh-package.json5`
