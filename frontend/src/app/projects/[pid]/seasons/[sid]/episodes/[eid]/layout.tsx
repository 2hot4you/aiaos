'use client';

import { useParams, usePathname, useRouter } from 'next/navigation';
import { Layout, Menu, Typography, Avatar, Dropdown } from 'antd';
import { FileTextOutlined, UserOutlined, VideoCameraOutlined, ExportOutlined, BulbOutlined, ArrowLeftOutlined, LogoutOutlined, SettingOutlined } from '@ant-design/icons';
import { useAuthStore } from '@/stores/authStore';

const { Sider, Content } = Layout;
const { Text } = Typography;

export default function WorkspaceLayout({ children }: { children: React.ReactNode }) {
  const params = useParams();
  const pathname = usePathname();
  const router = useRouter();
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);

  const pid = params.pid as string;
  const sid = params.sid as string;
  const eid = params.eid as string;
  const basePath = `/projects/${pid}/seasons/${sid}/episodes/${eid}`;

  const currentKey = pathname.includes('/script') ? 'script'
    : pathname.includes('/assets') ? 'assets'
    : pathname.includes('/director') ? 'director'
    : pathname.includes('/export') ? 'export'
    : pathname.includes('/prompts') ? 'prompts'
    : 'script';

  const menuItems = [
    { key: 'script', icon: <FileTextOutlined />, label: '剧本与故事' },
    { key: 'assets', icon: <UserOutlined />, label: '角色与场景' },
    { key: 'director', icon: <VideoCameraOutlined />, label: '导演工作台' },
    { key: 'export', icon: <ExportOutlined />, label: '成片与导出' },
    { key: 'prompts', icon: <BulbOutlined />, label: '提示词管理' },
  ];

  const userMenuItems = [
    { key: 'admin', icon: <SettingOutlined />, label: '管理后台', onClick: () => router.push('/admin/users') },
    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: logout },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider width={240} style={{ background: '#121828', borderRight: '1px solid #1E293B', display: 'flex', flexDirection: 'column' }}>
        <div style={{ padding: '16px 20px', borderBottom: '1px solid #1E293B' }}>
          <div
            style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer', color: '#94A3B8', marginBottom: 12, fontSize: 13 }}
            onClick={() => router.push(`/projects/${pid}`)}
          >
            <ArrowLeftOutlined /> 返回项目
          </div>
          <Text style={{ color: '#F1F5F9', fontWeight: 600, fontSize: 15, fontFamily: "'Space Grotesk', sans-serif" }}>
            工作台
          </Text>
          <br />
          <Text style={{ color: '#64748B', fontSize: 12 }}>S{sid?.slice(-2) || '?'} · E{eid?.slice(-2) || '?'}</Text>
        </div>

        <Menu
          mode="inline"
          selectedKeys={[currentKey]}
          items={menuItems}
          onClick={({ key }) => router.push(`${basePath}/${key}`)}
          style={{ background: 'transparent', borderRight: 'none', flex: 1, marginTop: 8 }}
          theme="dark"
        />

        <div style={{ padding: '16px 20px', borderTop: '1px solid #1E293B' }}>
          <Dropdown menu={{ items: user?.role === 'admin' ? userMenuItems : [userMenuItems[1]] }} trigger={['click']} placement="topLeft">
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, cursor: 'pointer' }}>
              <Avatar size={32} style={{ background: '#6366F1' }}>{user?.display_name?.[0] || 'U'}</Avatar>
              <div>
                <Text style={{ color: '#F1F5F9', fontSize: 13, display: 'block' }}>{user?.display_name}</Text>
                <Text style={{ color: '#64748B', fontSize: 11 }}>{user?.role === 'admin' ? '管理员' : '用户'}</Text>
              </div>
            </div>
          </Dropdown>
        </div>
      </Sider>
      <Content style={{ background: '#0A0E1A', overflow: 'auto' }}>
        {children}
      </Content>
    </Layout>
  );
}
