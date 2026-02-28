# UI Migration: Android View/Compose → ArkUI

本文档覆盖 Android UI 代码到 OpenHarmony ArkUI 声明式 UI 的迁移方法和常见模式。

## Table of Contents

- [核心概念对比](#核心概念对比)
- [Layout 迁移](#layout-迁移)
- [列表与滚动](#列表与滚动)
- [State 管理](#state-管理)
- [Navigation 迁移](#navigation-迁移)
- [动画迁移](#动画迁移)
- [自定义 View → 自定义组件](#自定义-view--自定义组件)
- [Resource 迁移](#resource-迁移)
- [常见 UI Pattern 对照](#常见-ui-pattern-对照)

---

## 核心概念对比

| Android | ArkUI (OpenHarmony) | 说明 |
|---------|-------------------|------|
| XML Layout + Activity | `@Component` struct + `build()` | 声明式，无 XML |
| View 继承体系 | 组件组合 (Composition) | 无继承，用组合 |
| `findViewById` | 不需要，声明式绑定 | |
| `setOnClickListener` | `.onClick(() => {})` | 链式调用 |
| `RecyclerView.Adapter` | `ForEach` / `LazyForEach` | |
| DP / SP | vp / fp | 虚拟像素 / 字体像素 |
| Theme / Style | `@Styles` / `@Extend` | |
| XML drawable | 组件属性直接设置 | `.backgroundColor().borderRadius()` |

### Android View vs ArkUI Component

```java
// Android View
public class MyView extends LinearLayout {
    private TextView title;
    private Button action;

    public MyView(Context context) {
        super(context);
        inflate(context, R.layout.my_view, this);
        title = findViewById(R.id.title);
        action = findViewById(R.id.action);
        action.setOnClickListener(v -> handleClick());
    }
}
```

```typescript
// ArkUI Component
@Component
struct MyView {
  @Prop title: string = ''
  private handleClick = () => {}

  build() {
    Column() {
      Text(this.title)
        .fontSize(16)
      Button('Action')
        .onClick(() => this.handleClick())
    }
  }
}
```

---

## Layout 迁移

### LinearLayout → Column / Row

```xml
<!-- Android vertical LinearLayout -->
<LinearLayout android:orientation="vertical"
    android:layout_width="match_parent"
    android:layout_height="wrap_content"
    android:padding="16dp">
    <TextView android:text="Title" />
    <TextView android:text="Subtitle" />
</LinearLayout>
```

```typescript
// ArkUI
Column() {
  Text('Title')
  Text('Subtitle')
}
.width('100%')
.padding(16)
```

### RelativeLayout → RelativeContainer

```typescript
RelativeContainer() {
  Text('Center')
    .id('center')
    .alignRules({
      center: { anchor: '__container__', align: VerticalAlign.Center },
      middle: { anchor: '__container__', align: HorizontalAlign.Center }
    })
}
```

### FrameLayout → Stack

```typescript
Stack({ alignContent: Alignment.BottomEnd }) {
  Image($r('app.media.bg'))
    .width('100%')
  Text('Overlay')
    .padding(8)
}
```

### weight (LinearLayout) → layoutWeight

```xml
<LinearLayout android:orientation="horizontal">
    <View android:layout_weight="1" />
    <View android:layout_weight="2" />
</LinearLayout>
```

```typescript
Row() {
  Text('1').layoutWeight(1)
  Text('2').layoutWeight(2)
}
```

---

## 列表与滚动

### RecyclerView → List + ForEach/LazyForEach

```java
// Android
RecyclerView recyclerView = findViewById(R.id.list);
recyclerView.setLayoutManager(new LinearLayoutManager(this));
recyclerView.setAdapter(new MyAdapter(dataList));
```

```typescript
// ArkUI - 小数据量
List() {
  ForEach(this.dataList, (item: MyItem) => {
    ListItem() {
      Text(item.name).fontSize(16)
    }
  }, (item: MyItem) => item.id.toString())
}

// ArkUI - 大数据量 (LazyForEach)
List() {
  LazyForEach(this.dataSource, (item: MyItem) => {
    ListItem() {
      Text(item.name).fontSize(16)
    }
  }, (item: MyItem) => item.id.toString())
}
```

### LazyForEach DataSource 实现

```typescript
class MyDataSource implements IDataSource {
  private data: MyItem[] = []

  totalCount(): number { return this.data.length }
  getData(index: number): MyItem { return this.data[index] }

  registerDataChangeListener(listener: DataChangeListener): void { /* ... */ }
  unregisterDataChangeListener(listener: DataChangeListener): void { /* ... */ }
}
```

### GridView → Grid

```typescript
Grid() {
  ForEach(this.items, (item: string) => {
    GridItem() {
      Text(item)
    }
  })
}
.columnsTemplate('1fr 1fr 1fr')  // 3 columns
.rowsGap(10)
.columnsGap(10)
```

---

## State 管理

### Android → ArkUI State 装饰器

| Android 模式 | ArkUI 装饰器 | 作用域 |
|-------------|------------|--------|
| `private field` | `@State` | 组件内状态 |
| constructor param (只读) | `@Prop` | 父到子单向同步 |
| callback / interface | `@Link` | 父子双向同步 |
| ViewModel + LiveData | `@Provide` / `@Consume` | 跨层级传递 |
| SharedPreferences | `AppStorage` / `PersistentStorage` | 全局 / 持久化 |
| EventBus | `@Watch` | 状态变化监听 |

### ViewModel 迁移模式

```java
// Android ViewModel
public class UserViewModel extends ViewModel {
    private MutableLiveData<User> user = new MutableLiveData<>();
    public LiveData<User> getUser() { return user; }
    public void loadUser(int id) {
        userRepo.getUser(id, result -> user.postValue(result));
    }
}
```

```typescript
// ArkUI - 使用 @State 管理
@Component
struct UserPage {
  @State user: User | null = null

  async loadUser(id: number) {
    this.user = await UserRepository.getUser(id)
    // @State 自动触发 UI 更新
  }

  build() {
    Column() {
      if (this.user) {
        Text(this.user.name)
      } else {
        LoadingIndicator()
      }
    }
  }
}
```

---

## Navigation 迁移

### Android Navigation → OH Navigation

```java
// Android - Intent navigation
Intent intent = new Intent(this, DetailActivity.class);
intent.putExtra("id", itemId);
startActivity(intent);
```

```typescript
// OpenHarmony - 页面内导航 (Router)
import router from '@ohos.router'
router.pushUrl({
  url: 'pages/Detail',
  params: { id: itemId }
})

// OpenHarmony - Navigation 组件 (推荐)
Navigation(this.navPathStack) {
  // root content
}
.navDestination(this.pageMap)

// 跳转
this.navPathStack.pushPath({ name: 'detail', param: { id: itemId } })
```

### Fragment Navigation → NavDestination

```typescript
// 路由表
@Builder
pageMap(name: string) {
  if (name === 'detail') {
    DetailPage()
  } else if (name === 'settings') {
    SettingsPage()
  }
}
```

---

## 动画迁移

### Android Animation → ArkUI Animation

| Android | ArkUI | Notes |
|---------|-------|-------|
| `ObjectAnimator` | `.animation()` 属性动画 | |
| `ValueAnimator` | `animateTo()` | |
| `TransitionManager` | `transition()` + `TransitionEffect` | |
| `MotionLayout` | 显式动画 `animateTo()` | |
| `Lottie` | `@ohos/lottie` | ohpm 第三方包 |

```typescript
// 属性动画
Text('Hello')
  .opacity(this.show ? 1 : 0)
  .animation({ duration: 300, curve: Curve.EaseInOut })

// 显式动画
animateTo({ duration: 500 }, () => {
  this.translateX = 200
})
```

---

## 自定义 View → 自定义组件

### Android 自定义 View 迁移步骤

1. 将 `onDraw(Canvas)` → ArkUI `Canvas` 组件 + `CanvasRenderingContext2D`
2. 将 `onMeasure/onLayout` → ArkUI 布局属性 (`.width()`, `.height()` 等) 
3. 将 attrs.xml 自定义属性 → `@Prop` / `@Link` 参数
4. 将 Touch 事件 → `.onTouch()` / `.gesture()` 手势

```java
// Android Custom View
public class CircleView extends View {
    @Override protected void onDraw(Canvas canvas) {
        Paint paint = new Paint();
        paint.setColor(Color.RED);
        canvas.drawCircle(getWidth()/2f, getHeight()/2f, 100f, paint);
    }
}
```

```typescript
// ArkUI Custom Drawing
@Component
struct CircleView {
  private settings: RenderingContextSettings = new RenderingContextSettings(true)
  private ctx: CanvasRenderingContext2D = new CanvasRenderingContext2D(this.settings)

  build() {
    Canvas(this.ctx)
      .width(200)
      .height(200)
      .onReady(() => {
        this.ctx.fillStyle = '#FF0000'
        this.ctx.beginPath()
        this.ctx.arc(100, 100, 50, 0, Math.PI * 2)
        this.ctx.fill()
      })
  }
}
```

---

## Resource 迁移

### Android res/ → OpenHarmony resources/

| Android | OpenHarmony | 路径 |
|---------|-------------|------|
| `res/values/strings.xml` | `resources/base/element/string.json` | |
| `res/values/colors.xml` | `resources/base/element/color.json` | |
| `res/drawable/` | `resources/base/media/` | |
| `res/layout/` | 无，声明式 UI 不需要 | |
| `res/raw/` | `resources/rawfile/` | |
| `res/mipmap/` | `resources/base/media/` | |

### 资源引用方式

```java
// Android
getString(R.string.app_name)
getDrawable(R.drawable.icon)
getColor(R.color.primary)
```

```typescript
// OpenHarmony
$r('app.string.app_name')     // 字符串
$r('app.media.icon')           // 图片
$r('app.color.primary')        // 颜色
$rawfile('data.json')          // 原始文件
```

---

## 常见 UI Pattern 对照

### Toolbar/ActionBar → Title + Menus

```typescript
@Entry
@Component
struct MyPage {
  build() {
    Navigation() {
      // page content
    }
    .title('Page Title')
    .menus([
      { value: 'Settings', icon: $r('app.media.settings'),
        action: () => { /* handle */ } }
    ])
  }
}
```

### BottomNavigationView → Tabs (Bottom)

```typescript
Tabs({ barPosition: BarPosition.End }) {
  TabContent() { HomePage() }
    .tabBar(this.tabBuilder('Home', 0, $r('app.media.home')))
  TabContent() { ProfilePage() }
    .tabBar(this.tabBuilder('Profile', 1, $r('app.media.profile')))
}
```

### FloatingActionButton → 自定义定位 Button

```typescript
Stack({ alignContent: Alignment.BottomEnd }) {
  // Page content
  List() { /* ... */ }

  Button({ type: ButtonType.Circle }) {
    Image($r('app.media.add')).width(24)
  }
  .width(56).height(56)
  .margin({ right: 16, bottom: 16 })
  .onClick(() => { /* handle */ })
}
```

### Pull-to-Refresh → Refresh 组件

```typescript
Refresh({ refreshing: $$this.isRefreshing }) {
  List() {
    ForEach(this.data, (item: string) => {
      ListItem() { Text(item) }
    })
  }
}
.onRefreshing(async () => {
  await this.loadData()
  this.isRefreshing = false
})
```
