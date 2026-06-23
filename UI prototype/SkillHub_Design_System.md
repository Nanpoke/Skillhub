# SkillHub 设计系统文档

> 赛博朋克风格设计规范 - 包含配色、组件样式、动画效果的完整指南

---

## 配色方案

### 背景色
```css
--bg-primary:   #0a0a0f;    /* 主背景 */
--bg-secondary: #12121a;    /* 面板底色 */
--bg-tertiary:  #1a1a24;    /* 卡片底色 */
```

### 边框色
```css
--border-subtle:  rgba(255, 255, 255, 0.06);
--border-default: rgba(255, 255, 255, 0.12);
--border-hover:   rgba(0, 212, 170, 0.4);
```

### 强调色
```css
--accent-primary:   #00d4aa;   /* 青绿 - 主色 */
--accent-secondary: #a855f7;   /* 紫色 - 次色 */
```

### 语义色
```css
--color-orange:  #f97316;    /* Claude Code */
--color-blue:    #3b82f6;    /* OpenCode */
--color-purple:  #a855f7;    /* Cursor */
--color-green:   #10b981;    /* CodeBuddy */
--color-pink:    #ec4899;    /* Trae */
--color-red:     #ef4444;    /* 删除/错误 */
```

### 文字色
```css
--text-white: #ffffff;
--text-gray:  #e5e7eb;
--text-muted: #9ca3af;
--text-dark:  #6b7280;
```

---

## 字体

```css
--font-sans: 'Inter', system-ui, sans-serif;
--font-mono: 'JetBrains Mono', monospace;
```

---

## 核心组件

### 卡片（渐变边框方案）

```css
.skill-card {
    position: relative;
    background: rgba(26, 26, 36, 0.6);
    backdrop-filter: blur(20px);
    border-radius: 16px;
    padding: 20px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* 渐变边框 */
.skill-card::before {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: inherit;
    padding: 1px;
    background: linear-gradient(135deg,
        rgba(255, 255, 255, 0.12) 0%,
        rgba(255, 255, 255, 0.04) 50%,
        rgba(255, 255, 255, 0.12) 100%
    );
    -webkit-mask: linear-gradient(#fff 0 0) content-box,
                  linear-gradient(#fff 0 0);
    -webkit-mask-composite: xor;
    mask-composite: exclude;
    opacity: 0.8;
    transition: opacity 0.3s, background 0.3s;
    pointer-events: none;
}

/* Hover 状态 */
.skill-card:hover {
    transform: translateY(-4px);
}

.skill-card:hover::before {
    opacity: 1;
    background: linear-gradient(135deg,
        rgba(0, 212, 170, 0.4) 0%,
        rgba(168, 85, 247, 0.3) 100%
    );
}
```

### 玻璃态面板

```css
.glass-panel {
    background: rgba(18, 18, 26, 0.8);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(255, 255, 255, 0.05);
    border-radius: 16px;
}
```

---

## 按钮

### 主要按钮

```css
.btn-primary {
    background: linear-gradient(135deg, #00d4aa, #a855f7);
    color: #0a0a0f;
    padding: 12px 32px;
    border-radius: 12px;
    font-weight: 600;
    border: none;
    cursor: pointer;
    box-shadow: 0 4px 16px rgba(0, 212, 170, 0.25);
    transition: all 0.3s ease;
}

.btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 24px rgba(0, 212, 170, 0.4),
                0 0 30px rgba(0, 212, 170, 0.2);
}

.btn-primary:active {
    transform: translateY(0) scale(0.98);
}

.btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
}
```

### 次要按钮

```css
.btn-secondary {
    background: transparent;
    border: 1px solid rgba(0, 212, 170, 0.4);
    color: #00d4aa;
    padding: 10px 20px;
    border-radius: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.3s ease;
}

.btn-secondary:hover {
    background: rgba(0, 212, 170, 0.1);
    border-color: rgba(0, 212, 170, 0.6);
    box-shadow: 0 0 20px rgba(0, 212, 170, 0.15);
}
```

### 幽灵按钮

```css
.btn-ghost {
    background: transparent;
    border: 1px solid transparent;
    color: #9ca3af;
    padding: 8px 16px;
    border-radius: 10px;
    cursor: pointer;
    transition: all 0.2s ease;
}

.btn-ghost:hover {
    background: rgba(255, 255, 255, 0.05);
    color: #e5e7eb;
}
```

### 图标按钮

```css
.btn-icon {
    width: 40px;
    height: 40px;
    background: rgba(18, 18, 26, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    color: #9ca3af;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s ease;
}

.btn-icon:hover {
    border-color: rgba(0, 212, 170, 0.4);
    color: #00d4aa;
    background: rgba(0, 212, 170, 0.08);
    box-shadow: 0 0 16px rgba(0, 212, 170, 0.1);
}

.btn-icon:hover i {
    transform: scale(1.1);
}

/* 删除按钮 */
.btn-icon.delete:hover {
    border-color: rgba(239, 68, 68, 0.4);
    color: #ef4444;
    background: rgba(239, 68, 68, 0.08);
    box-shadow: 0 0 16px rgba(239, 68, 68, 0.1);
}
```

---

## 表单组件

### Toggle 开关

```css
.toggle-switch {
    position: relative;
    width: 44px;
    height: 24px;
    background: rgba(30, 30, 46, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.3s ease;
}

.toggle-switch:hover {
    border-color: rgba(0, 212, 170, 0.3);
}

.toggle-switch.active {
    background: linear-gradient(135deg, #00d4aa, #00b894);
    border-color: transparent;
    box-shadow: 0 0 12px rgba(0, 212, 170, 0.4);
}

.toggle-switch::after {
    content: '';
    position: absolute;
    top: 2px;
    left: 2px;
    width: 20px;
    height: 20px;
    background: white;
    border-radius: 50%;
    transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.toggle-switch.active::after {
    transform: translateX(20px);
}
```

### 复选框

```css
.cyber-checkbox {
    appearance: none;
    width: 18px;
    height: 18px;
    background: rgba(18, 18, 26, 0.8);
    border: 1.5px solid rgba(255, 255, 255, 0.15);
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.2s ease;
    position: relative;
}

.cyber-checkbox:hover {
    border-color: rgba(0, 212, 170, 0.4);
    background: rgba(0, 212, 170, 0.05);
}

.cyber-checkbox:checked {
    background: #00d4aa;
    border-color: #00d4aa;
    box-shadow: 0 0 12px rgba(0, 212, 170, 0.4);
}

.cyber-checkbox:checked::after {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%) rotate(45deg);
    width: 4px;
    height: 8px;
    border: solid #0a0a0f;
    border-width: 0 2px 2px 0;
}
```

### 单选按钮

```css
.cyber-radio {
    appearance: none;
    width: 20px;
    height: 20px;
    background: rgba(18, 18, 26, 0.8);
    border: 2px solid rgba(255, 255, 255, 0.15);
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.2s ease;
    position: relative;
}

.cyber-radio:hover {
    border-color: rgba(0, 212, 170, 0.4);
    background: rgba(0, 212, 170, 0.05);
}

.cyber-radio:checked {
    border-color: #00d4aa;
    background: #00d4aa;
    box-shadow: 0 0 12px rgba(0, 212, 170, 0.4);
}

.cyber-radio:checked::after {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 6px;
    height: 6px;
    background: #0a0a0f;
    border-radius: 50%;
}
```

### 输入框

```css
.input-field {
    width: 100%;
    padding: 12px 16px;
    background: rgba(10, 10, 15, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    color: #e5e7eb;
    font-size: 14px;
    transition: all 0.2s ease;
}

.input-field:hover {
    border-color: rgba(255, 255, 255, 0.15);
}

.input-field:focus {
    outline: none;
    border-color: #00d4aa;
    box-shadow: 0 0 0 3px rgba(0, 212, 170, 0.1),
                0 0 20px rgba(0, 212, 170, 0.1);
}

.input-field::placeholder {
    color: #6b7280;
}
```

---

## 标签

```css
.tag {
    display: inline-flex;
    align-items: center;
    padding: 6px 12px;
    background: rgba(0, 212, 170, 0.1);
    border: 1px solid rgba(0, 212, 170, 0.2);
    border-radius: 8px;
    font-size: 12px;
    color: #00d4aa;
}

/* 可交互标签 */
.tag.interactive {
    cursor: pointer;
    transition: all 0.2s ease;
}

.tag.interactive:hover {
    background: rgba(0, 212, 170, 0.15);
    border-color: rgba(0, 212, 170, 0.4);
}

/* 可删除标签 */
.tag .tag-remove {
    margin-left: 6px;
    opacity: 0.6;
    cursor: pointer;
    transition: all 0.2s;
}

.tag .tag-remove:hover {
    opacity: 1;
    color: #ef4444;
}
```

---

## 动画

### 页面加载

```css
@keyframes slideUp {
    from { opacity: 0; transform: translateY(20px); }
    to   { opacity: 1; transform: translateY(0); }
}

.animate-slide-up {
    animation: slideUp 0.5s ease-out;
}
```

### 发光脉冲

```css
@keyframes glow {
    0%   { box-shadow: 0 0 20px rgba(0, 212, 170, 0.1); }
    100% { box-shadow: 0 0 40px rgba(0, 212, 170, 0.3),
                      0 0 60px rgba(168, 85, 247, 0.1); }
}

.animate-glow {
    animation: glow 3s ease-in-out infinite alternate;
}
```

### 背景光晕浮动

```css
@keyframes float {
    0%, 100% { transform: translate(0, 0); }
    50%      { transform: translate(30px, -30px); }
}

.bg-glow {
    position: fixed;
    width: 600px;
    height: 600px;
    border-radius: 50%;
    filter: blur(100px);
    opacity: 0.15;
    pointer-events: none;
    animation: float 8s ease-in-out infinite;
}
```

### 展开面板

```css
.expand-panel {
    max-height: 0;
    overflow: hidden;
    transition: max-height 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

.expand-panel.open {
    max-height: 500px;
}
```

### 交错动画

```css
.stagger-children > * {
    opacity: 0;
    animation: slideUp 0.4s ease-out forwards;
}

.stagger-children > *:nth-child(1) { animation-delay: 0ms; }
.stagger-children > *:nth-child(2) { animation-delay: 50ms; }
.stagger-children > *:nth-child(3) { animation-delay: 100ms; }
.stagger-children > *:nth-child(4) { animation-delay: 150ms; }
.stagger-children > *:nth-child(5) { animation-delay: 200ms; }
```

---

## 圆角

```css
--radius-sm:   8px;   /* 标签、小按钮 */
--radius-md:   12px;  /* 按钮、输入框 */
--radius-lg:   16px;  /* 卡片、面板 */
--radius-xl:   24px;  /* 大面板、模态框 */
```

---

## 滚动条

```css
::-webkit-scrollbar {
    width: 8px;
    height: 8px;
}

::-webkit-scrollbar-track {
    background: #0a0a0f;
}

::-webkit-scrollbar-thumb {
    background: #1e1e2e;
    border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
    background: #00d4aa;
}
```

---

## 工具品牌色

| 工具 | 标识 | 颜色 |
|------|------|------|
| Claude Code | C | `#f97316` 橙色 |
| OpenCode | O | `#3b82f6` 蓝色 |
| Cursor | Cu | `#a855f7` 紫色 |
| CodeBuddy | CB | `#10b981` 绿色 |
| Trae | T | `#ec4899` 粉色 |

---

## 分类图标渐变

```css
.cat-frontend { background: linear-gradient(135deg, rgba(249, 115, 22, 0.2), rgba(239, 68, 68, 0.2)); }
.cat-backend  { background: linear-gradient(135deg, rgba(34, 197, 94, 0.2), rgba(16, 185, 129, 0.2)); }
.cat-design   { background: linear-gradient(135deg, rgba(168, 85, 247, 0.2), rgba(236, 72, 153, 0.2)); }
.cat-devops   { background: linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(6, 182, 212, 0.2)); }
```

---

## 发光效果

```css
/* 主色发光 */
.glow-accent {
    box-shadow: 0 0 20px rgba(0, 212, 170, 0.3);
}

/* 组合发光 */
.glow-mixed {
    box-shadow: 0 0 20px rgba(0, 212, 170, 0.3),
                0 0 40px rgba(168, 85, 247, 0.1);
}

/* 红色发光 */
.glow-red {
    box-shadow: 0 0 20px rgba(239, 68, 68, 0.3);
}
```

---

## 动画曲线

```css
--ease-out:     cubic-bezier(0, 0, 0.2, 1);
--ease-in-out:  cubic-bezier(0.4, 0, 0.6, 1);
--spring:       cubic-bezier(0.175, 0.885, 0.32, 1.1);
--bounce:       cubic-bezier(0.34, 1.56, 0.64, 1);
```

---

## Vue 组件示例

```vue
<template>
  <div class="skill-card" @click="$emit('click')">
    <slot />
  </div>
</template>

<style scoped>
.skill-card {
  position: relative;
  background: rgba(26, 26, 36, 0.6);
  backdrop-filter: blur(20px);
  border-radius: 16px;
  padding: 20px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.skill-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  padding: 1px;
  background: linear-gradient(135deg,
    rgba(255, 255, 255, 0.12) 0%,
    rgba(255, 255, 255, 0.04) 50%,
    rgba(255, 255, 255, 0.12) 100%
  );
  -webkit-mask: linear-gradient(#fff 0 0) content-box,
                linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  opacity: 0.8;
  transition: opacity 0.3s, background 0.3s;
  pointer-events: none;
}

.skill-card:hover {
  transform: translateY(-4px);
}

.skill-card:hover::before {
  opacity: 1;
  background: linear-gradient(135deg,
    rgba(0, 212, 170, 0.4) 0%,
    rgba(168, 85, 247, 0.3) 100%
  );
}
</style>
```

---

## Tailwind 配置

```javascript
// tailwind.config.js
module.exports = {
  theme: {
    extend: {
      colors: {
        cyber: {
          dark: '#0a0a0f',
          panel: '#12121a',
          border: '#1e1e2e',
          accent: '#00d4aa',
          accent2: '#a855f7',
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      animation: {
        'glow': 'glow 3s ease-in-out infinite alternate',
        'slide-up': 'slideUp 0.5s ease-out',
        'float': 'float 8s ease-in-out infinite',
      },
      keyframes: {
        glow: {
          '0%': { boxShadow: '0 0 20px rgba(0, 212, 170, 0.1)' },
          '100%': { boxShadow: '0 0 40px rgba(0, 212, 170, 0.3), 0 0 60px rgba(168, 85, 247, 0.1)' },
        },
        slideUp: {
          '0%': { transform: 'translateY(20px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        float: {
          '0%, 100%': { transform: 'translate(0, 0)' },
          '50%': { transform: 'translate(30px, -30px)' },
        },
      },
    }
  }
}
```

---

*版本: v2.0 | 更新: 2026-06-21*
