'use client';

import { ConfigProvider, App as AntApp } from 'antd';
import theme from '@/theme/antdTheme';
import AuthGuard from '@/components/auth/AuthGuard';

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ConfigProvider theme={{ ...theme, algorithm: undefined }}>
      <AntApp>
        <AuthGuard>{children}</AuthGuard>
      </AntApp>
    </ConfigProvider>
  );
}
