'use client';

import { useState } from 'react';
import { Form, Input, Button, Card, message, Typography } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';

const { Title, Text } = Typography;

export default function LoginPage() {
  const [loading, setLoading] = useState(false);
  const login = useAuthStore((s) => s.login);
  const router = useRouter();

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await login(values.username, values.password);
      message.success('登录成功');
      router.push('/projects');
    } catch (err: any) {
      message.error(err.message || '登录失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'linear-gradient(135deg, #0A0E1A 0%, #121828 50%, #0A0E1A 100%)',
    }}>
      <Card
        style={{
          width: 400,
          background: '#121828',
          border: '1px solid #1E293B',
          borderRadius: 12,
        }}
        styles={{ body: { padding: '40px 32px' } }}
      >
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <div style={{
            width: 56,
            height: 56,
            background: 'linear-gradient(135deg, #6366F1, #8B5CF6)',
            borderRadius: 14,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            margin: '0 auto 16px',
            fontSize: 24,
          }}>
            🎬
          </div>
          <Title level={3} style={{ color: '#F1F5F9', margin: 0, fontFamily: "'Space Grotesk', sans-serif" }}>
            AIAOS
          </Title>
          <Text style={{ color: '#94A3B8', fontSize: 13 }}>AI 漫剧生成平台</Text>
        </div>

        <Form onFinish={onFinish} layout="vertical" size="large">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input prefix={<UserOutlined style={{ color: '#94A3B8' }} />} placeholder="用户名" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined style={{ color: '#94A3B8' }} />} placeholder="密码" />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0 }}>
            <Button type="primary" htmlType="submit" loading={loading} block style={{ height: 44, borderRadius: 8, fontWeight: 500 }}>
              登录
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center', marginTop: 20 }}>
          <Text style={{ color: '#64748B', fontSize: 12 }}>忘记密码请联系系统管理员</Text>
        </div>
      </Card>
    </div>
  );
}
