# UI 迁移：Android View/Compose → ArkUI

本 skill 覆盖 Android UI 代码到 OpenHarmony ArkUI 声明式 UI 的迁移。

---

## 核心概念对比

| Android | ArkUI (OpenHarmony) | 说明 |
|---------|-------------------|------|
| XML Layout + Activity | `@Component` struct + `build()` | 声明式，无 XML |
| View 继承体系 | 组件组合 (Composition) | 无继承，用组合 |
| `findViewById` | 不需要，声明式绑定 | |
| `setOnClickListener` | `.onClick(() => {})` | 链式调用 |
| DP / SP | vp / fp | 虚拟像素 / 字体像素 |

---

## 组件映射速查

| Android | ArkUI | 示例 |
|---------|-------|------|
| `TextView` | `Text()` | `Text('hello').fontSize(16)` |
| `Button` | `Button()` | `Button('Click').onClick(()=>{})` |
| `ImageView` | `Image()` | `Image($r('app.media.icon'))` |
| `EditText` | `TextInput()` / `TextArea()` | |
| `RecyclerView` | `List` + `ForEach` | |
| `LinearLayout` (垂直) | `Column` | |
| `LinearLayout` (水平) | `Row` | |
| `FrameLayout` | `Stack` | |
| `ScrollView` | `Scroll` | |
| `ViewPager` | `Swiper` | |
| `TabLayout` | `Tabs` + `TabContent` | |
| `ProgressBar` | `Progress` | |
| `Switch` | `Toggle` | |
| `CheckBox` | `Checkbox` | |
| `WebView` | `Web` | |

### Jetpack Compose → ArkUI

| Compose | ArkUI |
|---------|-------|
| `@Composable` | `@Component` + `build()` |
| `remember { mutableStateOf() }` | `@State` |
| `LazyColumn` | `List` + `LazyForEach` |
| `Modifier` | 链式属性 `.width().height()` |
| `Column` | `Column` |
| `Row` | `Row` |
| `Box` | `Stack` |
| `NavHost` | `Navigation` |

---

## State 管理

| Android 模式 | ArkUI 装饰器 | 作用域 |
|-------------|------------|--------|
| `private field` | `@State` | 组件内状态 |
| constructor param (只读) | `@Prop` | 父到子单向同步 |
| callback / interface | `@Link` | 父子双向同步 |
| ViewModel + LiveData | `@Provide` / `@Consume` | 跨层级传递 |
| SharedPreferences | `AppStorage` / `PersistentStorage` | 全局 / 持久化 |
| EventBus | `@Watch` | 状态变化监听 |

---

## Layout 迁移示例

```xml
<!-- Android -->
<LinearLayout android:orientation="vertical"
    android:padding="16dp">
    <TextView android:text="Title" />
    <Button android:text="Click" />
</LinearLayout>
```

```typescript
// ArkUI
Column() {
  Text('Title')
  Button('Click').onClick(() => {})
}
.padding(16)
```

---

## 列表迁移

```typescript
// 小数据量 - ForEach
List() {
  ForEach(this.dataList, (item: MyItem) => {
    ListItem() {
      Text(item.name).fontSize(16)
    }
  }, (item: MyItem) => item.id.toString())
}

// 大数据量 - LazyForEach
List() {
  LazyForEach(this.dataSource, (item: MyItem) => {
    ListItem() {
      Text(item.name).fontSize(16)
    }
  }, (item: MyItem) => item.id.toString())
}
```

---

## Navigation 迁移

```java
// Android
Intent intent = new Intent(this, DetailActivity.class);
intent.putExtra("id", itemId);
startActivity(intent);
```

```typescript
// OpenHarmony - Router
import router from '@ohos.router';
router.pushUrl({
  url: 'pages/Detail',
  params: { id: itemId }
});

// OpenHarmony - Navigation 组件（推荐）
this.navPathStack.pushPath({ name: 'detail', param: { id: itemId } });
```

---

## 资源引用

| Android | OpenHarmony |
|---------|-------------|
| `getString(R.string.app_name)` | `$r('app.string.app_name')` |
| `getDrawable(R.drawable.icon)` | `$r('app.media.icon')` |
| `getColor(R.color.primary)` | `$r('app.color.primary')` |
| `raw/data.json` | `$rawfile('data.json')` |

---

## 常见 UI Pattern

### Pull-to-Refresh

```typescript
Refresh({ refreshing: $$this.isRefreshing }) {
  List() { ForEach(this.data, ...) }
}
.onRefreshing(async () => {
  await this.loadData();
  this.isRefreshing = false;
})
```

### FloatingActionButton

```typescript
Stack({ alignContent: Alignment.BottomEnd }) {
  List() { /* content */ }
  Button({ type: ButtonType.Circle }) {
    Image($r('app.media.add')).width(24)
  }
  .width(56).height(56)
  .margin({ right: 16, bottom: 16 })
}
```

### BottomNavigation

```typescript
Tabs({ barPosition: BarPosition.End }) {
  TabContent() { HomePage() }.tabBar('Home')
  TabContent() { ProfilePage() }.tabBar('Profile')
}
```

---

## 详细参考

- [references/ui-migration.md](../../references/ui-migration.md) — View/Compose → ArkUI 完整映射与示例
