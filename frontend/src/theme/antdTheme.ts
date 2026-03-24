import type { ThemeConfig } from 'antd';

const theme: ThemeConfig = {
  token: {
    colorPrimary: '#6366F1',
    colorSuccess: '#22C55E',
    colorWarning: '#F59E0B',
    colorError: '#EF4444',
    colorInfo: '#6366F1',
    colorBgBase: '#0A0E1A',
    colorBgContainer: '#121828',
    colorBgElevated: '#1E293B',
    colorBgLayout: '#0A0E1A',
    colorText: '#F1F5F9',
    colorTextSecondary: '#94A3B8',
    colorTextTertiary: '#64748B',
    colorTextQuaternary: '#475569',
    colorBorder: '#1E293B',
    colorBorderSecondary: '#334155',
    borderRadius: 8,
    fontFamily: "'DM Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
    fontSize: 14,
    controlHeight: 40,
  },
  components: {
    Layout: {
      headerBg: '#121828',
      siderBg: '#121828',
      bodyBg: '#0A0E1A',
    },
    Menu: {
      darkItemBg: '#121828',
      darkItemSelectedBg: '#1E293B',
      darkItemColor: '#94A3B8',
      darkItemSelectedColor: '#F1F5F9',
    },
    Card: {
      colorBgContainer: '#121828',
    },
    Table: {
      colorBgContainer: '#121828',
      headerBg: '#1E293B',
      rowHoverBg: '#1E293B',
    },
    Modal: {
      contentBg: '#121828',
      headerBg: '#121828',
    },
    Input: {
      colorBgContainer: '#1E293B',
      activeBorderColor: '#6366F1',
    },
    Select: {
      colorBgContainer: '#1E293B',
    },
    Button: {
      primaryShadow: 'none',
    },
  },
};

export default theme;
