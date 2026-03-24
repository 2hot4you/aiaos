'use client';

import { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, Input, Select, Switch, message, Typography, Space, Tag, Popconfirm } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, KeyOutlined } from '@ant-design/icons';
import api from '@/lib/api';
import type { User } from '@/types';

const { Title } = Typography;

export default function AdminUsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [resetModalOpen, setResetModalOpen] = useState(false);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [form] = Form.useForm();
  const [resetForm] = Form.useForm();

  const fetchUsers = async () => {
    setLoading(true);
    try {
      const res: any = await api.get('/api/v1/admin/users');
      setUsers(res.data.items || []);
    } catch (err: any) { message.error(err.message); }
    finally { setLoading(false); }
  };

  useEffect(() => { fetchUsers(); }, []);

  const handleCreate = async (values: any) => {
    try {
      await api.post('/api/v1/admin/users', values);
      message.success('用户创建成功');
      setModalOpen(false);
      form.resetFields();
      fetchUsers();
    } catch (err: any) { message.error(err.message); }
  };

  const handleDelete = async (id: string) => {
    try {
      await api.delete(`/api/v1/admin/users/${id}`);
      message.success('删除成功');
      fetchUsers();
    } catch (err: any) { message.error(err.message); }
  };

  const handleResetPassword = async (values: { new_password: string }) => {
    if (!currentUser) return;
    try {
      await api.post(`/api/v1/admin/users/${currentUser.id}/reset-password`, values);
      message.success('密码重置成功');
      setResetModalOpen(false);
      resetForm.resetFields();
    } catch (err: any) { message.error(err.message); }
  };

  const columns = [
    { title: '用户名', dataIndex: 'username', key: 'username' },
    { title: '显示名称', dataIndex: 'display_name', key: 'display_name' },
    { title: '角色', dataIndex: 'role', key: 'role', render: (role: string) => <Tag color={role === 'admin' ? 'purple' : 'blue'}>{role === 'admin' ? '管理员' : '用户'}</Tag> },
    { title: '状态', dataIndex: 'enabled', key: 'enabled', render: (enabled: boolean) => <Tag color={enabled ? 'green' : 'red'}>{enabled ? '启用' : '禁用'}</Tag> },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', render: (t: string) => new Date(t).toLocaleDateString('zh-CN') },
    {
      title: '操作', key: 'action', width: 200,
      render: (_: any, record: User) => (
        <Space>
          <Button size="small" icon={<KeyOutlined />} onClick={() => { setCurrentUser(record); setResetModalOpen(true); }}>重置密码</Button>
          <Popconfirm title="确认删除？" onConfirm={() => handleDelete(record.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: '32px 48px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={4} style={{ color: '#F1F5F9', margin: 0 }}>用户管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>创建用户</Button>
      </div>

      <Table columns={columns} dataSource={users} rowKey="id" loading={loading} style={{ background: '#121828', borderRadius: 8 }} />

      <Modal title="创建用户" open={modalOpen} onCancel={() => setModalOpen(false)} footer={null}>
        <Form form={form} onFinish={handleCreate} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="display_name" label="显示名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="password" label="初始密码" rules={[{ required: true, min: 8, message: '至少 8 位' }]}><Input.Password /></Form.Item>
          <Form.Item name="role" label="角色" initialValue="user"><Select options={[{ value: 'user', label: '用户' }, { value: 'admin', label: '管理员' }]} /></Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}><Space><Button onClick={() => setModalOpen(false)}>取消</Button><Button type="primary" htmlType="submit">创建</Button></Space></Form.Item>
        </Form>
      </Modal>

      <Modal title={`重置密码 - ${currentUser?.display_name}`} open={resetModalOpen} onCancel={() => setResetModalOpen(false)} footer={null}>
        <Form form={resetForm} onFinish={handleResetPassword} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="new_password" label="新密码" rules={[{ required: true, min: 8, message: '至少 8 位' }]}><Input.Password /></Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}><Space><Button onClick={() => setResetModalOpen(false)}>取消</Button><Button type="primary" htmlType="submit">确认重置</Button></Space></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
