'use client';

import { usePathname, useRouter } from 'next/navigation';
import { Layout, Menu, Typography, Button } from 'antd';
import { UserOutlined, ApiOutlined, SettingOutlined, ArrowLeftOutlined, ControlOutlined } from '@ant-design/icons';

const { Sider, Content } = Layout;
const { Title } = Typography;

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();

  const currentKey = pathname.includes('/users') ? 'users'
    : pathname.includes('/models') ? 'models'
    : pathname.includes('/settings') ? 'settings'
    : pathname.includes('/components') ? 'components'
    : 'users';

  const menuItems = [
    { key: 'users', icon: <UserOutlined />, label: '用户管理' },
    { key: 'models', icon: <ApiOutlined />, label: '模型管理' },
    { key: 'settings', icon: <SettingOutlined />, label: '系统设置' },
    { key: 'components', icon: <ControlOutlined />, label: '组件参数' },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider width={220} style={{ background: '#121828', borderRight: '1px solid #1E293B' }}>
        <div style={{ padding: '16px 20px', borderBottom: '1px solid #1E293B' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer', color: '#94A3B8', marginBottom: 12, fontSize: 13 }}
            onClick={() => router.push('/projects')}>
            <ArrowLeftOutlined /> 返回项目库
          </div>
          <Title level={5} style={{ color: '#F1F5F9', margin: 0, fontFamily: "'Space Grotesk', sans-serif" }}>
            管理后台
          </Title>
        </div>
        <Menu mode="inline" selectedKeys={[currentKey]} items={menuItems}
          onClick={({ key }) => router.push(`/admin/${key}`)}
          style={{ background: 'transparent', borderRight: 'none', marginTop: 8 }} theme="dark" />
      </Sider>
      <Content style={{ background: '#0A0E1A', overflow: 'auto' }}>
        {children}
      </Content>
    </Layout>
  );
}
