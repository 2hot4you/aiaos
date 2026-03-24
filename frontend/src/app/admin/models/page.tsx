'use client';

import { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, Input, Select, Switch, message, Typography, Space, Tag, Popconfirm } from 'antd';
import { PlusOutlined, DeleteOutlined, CheckCircleOutlined } from '@ant-design/icons';
import api from '@/lib/api';
import type { ModelConfig } from '@/types';

const { Title } = Typography;

export default function AdminModelsPage() {
  const [models, setModels] = useState<ModelConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();

  const fetchModels = async () => {
    setLoading(true);
    try {
      const res: any = await api.get('/api/v1/admin/models');
      setModels(res.data.items || res.data || []);
    } catch (err: any) { message.error(err.message); }
    finally { setLoading(false); }
  };

  useEffect(() => { fetchModels(); }, []);

  const handleCreate = async (values: any) => {
    try {
      await api.post('/api/v1/admin/models', values);
      message.success('模型创建成功');
      setModalOpen(false);
      form.resetFields();
      fetchModels();
    } catch (err: any) { message.error(err.message); }
  };

  const handleDelete = async (id: string) => {
    try {
      await api.delete(`/api/v1/admin/models/${id}`);
      message.success('删除成功');
      fetchModels();
    } catch (err: any) { message.error(err.message); }
  };

  const typeColors: Record<string, string> = { text: 'blue', image: 'green', video: 'purple' };
  const typeLabels: Record<string, string> = { text: '生文', image: '生图', video: '生视频' };

  const columns = [
    { title: '模型名称', dataIndex: 'name', key: 'name', render: (t: string) => <span style={{ fontWeight: 500 }}>{t}</span> },
    { title: '类型', dataIndex: 'model_type', key: 'model_type', render: (t: string) => <Tag color={typeColors[t]}>{typeLabels[t] || t}</Tag> },
    { title: 'Provider', dataIndex: 'provider', key: 'provider' },
    { title: '模型标识', dataIndex: 'model_identifier', key: 'model_identifier', render: (t: string) => <code style={{ color: '#94A3B8', fontSize: 12 }}>{t}</code> },
    { title: '默认', dataIndex: 'is_default', key: 'is_default', render: (v: boolean) => v ? <CheckCircleOutlined style={{ color: '#22C55E' }} /> : null },
    { title: '状态', dataIndex: 'enabled', key: 'enabled', render: (v: boolean) => <Tag color={v ? 'green' : 'red'}>{v ? '启用' : '禁用'}</Tag> },
    {
      title: '操作', key: 'action', width: 100,
      render: (_: any, record: ModelConfig) => (
        <Popconfirm title="确认删除？" onConfirm={() => handleDelete(record.id)}>
          <Button size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  return (
    <div style={{ padding: '32px 48px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={4} style={{ color: '#F1F5F9', margin: 0 }}>模型管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>添加模型</Button>
      </div>

      <Table columns={columns} dataSource={models} rowKey="id" loading={loading} style={{ background: '#121828', borderRadius: 8 }} />

      <Modal title="添加模型" open={modalOpen} onCancel={() => setModalOpen(false)} footer={null} width={520}>
        <Form form={form} onFinish={handleCreate} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="模型名称" rules={[{ required: true }]}><Input placeholder="GPT-5.4" /></Form.Item>
          <Form.Item name="model_type" label="模型类型" rules={[{ required: true }]}>
            <Select options={[{ value: 'text', label: '生文' }, { value: 'image', label: '生图' }, { value: 'video', label: '生视频' }]} />
          </Form.Item>
          <Form.Item name="provider" label="Provider" rules={[{ required: true }]}>
            <Select options={[{ value: 'openai', label: 'OpenAI' }, { value: 'google', label: 'Google' }, { value: 'sora', label: 'Sora' }, { value: 'custom', label: 'Custom' }]} />
          </Form.Item>
          <Form.Item name="api_endpoint" label="API Endpoint" rules={[{ required: true }]}><Input placeholder="https://api.openai.com/v1" /></Form.Item>
          <Form.Item name="api_key" label="API Key" rules={[{ required: true }]}><Input.Password placeholder="sk-..." /></Form.Item>
          <Form.Item name="model_identifier" label="模型标识" rules={[{ required: true }]}><Input placeholder="gpt-5.4" /></Form.Item>
          <Form.Item name="is_default" label="设为默认" valuePropName="checked"><Switch /></Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}><Space><Button onClick={() => setModalOpen(false)}>取消</Button><Button type="primary" htmlType="submit">添加</Button></Space></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
