// SkillHub Design System - Tailwind 配置
// 版本: v2.0 | 更新: 2024-01-15

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      // === 配色方案 ===
      colors: {
        // 背景色
        'cyber-dark': '#0a0a0f',
        'cyber-panel': '#12121a',
        'cyber-tertiary': '#1a1a24',

        // 边框色
        'cyber-border': '#1e1e2e',

        // 强调色
        'cyber-accent': '#00d4aa',
        'cyber-accent2': '#a855f7',

        // 语义色 - AI 工具品牌色
        'claude': '#f97316',    // Claude Code - 橙色
        'opencode': '#3b82f6',   // OpenCode - 蓝色
        'cursor': '#a855f7',     // Cursor - 紫色
        'codebuddy': '#10b981',  // CodeBuddy - 绿色
        'trae': '#ec4899',       // Trae - 粉色

        // 状态色
        'success': '#10b981',
        'warning': '#f97316',
        'error': '#ef4444',
        'info': '#3b82f6',

        // 文字色
        'text-white': '#ffffff',
        'text-gray': '#e5e7eb',
        'text-muted': '#9ca3af',
        'text-dark': '#6b7280',
      },

      // === 字体 ===
      fontFamily: {
        'sans': ['Inter', 'system-ui', 'sans-serif'],
        'mono': ['JetBrains Mono', 'monospace'],
      },

      // === 圆角 ===
      borderRadius: {
        'sm': '8px',   // 标签、小按钮
        'md': '12px',  // 按钮、输入框
        'lg': '16px',  // 卡片、面板
        'xl': '24px',  // 大面板、模态框
      },

      // === 动画 ===
      animation: {
        'glow': 'glow 3s ease-in-out infinite alternate',
        'slide-up': 'slideUp 0.5s ease-out',
        'float': 'float 8s ease-in-out infinite',
        'fade-in': 'fadeIn 0.3s ease-out',
        'pulse': 'pulse 2s ease-in-out infinite',
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
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        pulse: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.5' },
        },
      },

      // === 阴影 ===
      boxShadow: {
        'glow-accent': '0 0 20px rgba(0, 212, 170, 0.3)',
        'glow-mixed': '0 0 20px rgba(0, 212, 170, 0.3), 0 0 40px rgba(168, 85, 247, 0.1)',
        'glow-red': '0 0 20px rgba(239, 68, 68, 0.3)',
        'card-hover': '0 20px 40px rgba(0, 0, 0, 0.4), 0 0 30px rgba(0, 212, 170, 0.1)',
        'button-hover': '0 6px 24px rgba(0, 212, 170, 0.4), 0 0 30px rgba(0, 212, 170, 0.2)',
      },

      // === 模糊效果 ===
      backdropBlur: {
        'glass': '20px',
        'strong': '30px',
      },

      // === 过渡效果 ===
      transitionDuration: {
        'fast': '150ms',
        'normal': '300ms',
        'slow': '500ms',
      },

      transitionTimingFunction: {
        'ease-out': 'cubic-bezier(0, 0, 0.2, 1)',
        'ease-in-out': 'cubic-bezier(0.4, 0, 0.6, 1)',
        'spring': 'cubic-bezier(0.175, 0.885, 0.32, 1.1)',
        'bounce': 'cubic-bezier(0.34, 1.56, 0.64, 1)',
      },

      // === 背景渐变 ===
      backgroundImage: {
        'accent-gradient': 'linear-gradient(135deg, #00d4aa, #a855f7)',
        'panel-gradient': 'linear-gradient(135deg, rgba(255, 255, 255, 0.12) 0%, rgba(255, 255, 255, 0.04) 50%, rgba(255, 255, 255, 0.12) 100%)',
        'card-hover-gradient': 'linear-gradient(135deg, rgba(0, 212, 170, 0.4) 0%, rgba(168, 85, 247, 0.3) 100%)',
      },

      // === Z-index ===
      zIndex: {
        'base': '0',
        'dropdown': '100',
        'modal': '1000',
        'tooltip': '2000',
      },

      // === 间距 ===
      spacing: {
        'xs': '4px',
        'sm': '8px',
        'md': '16px',
        'lg': '24px',
        'xl': '32px',
        '2xl': '48px',
      },
    },
  },
  plugins: [],
};
