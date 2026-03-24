'use client';

import { useEffect, useState } from 'react';
import { Button, Card, Table, Modal, Form, Input, message, Typography, Space, Row, Col, Statistic, Tag } from 'antd';
import { PlusOutlined, FolderOpenOutlined, ProjectOutlined, ClockCircleOutlined } from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import api from '@/lib/api';
import type { Project } from '@/types';

const { Title } = Typography;

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [creating, setCreating] = useState(false);
  const [form] = Form.useForm();
  const router = useRouter();

  const fetchProjects = async () => {
    setLoading(true);
    try {
      const res: any = await api.get('/api/v1/projects');
      setProjects(res.data.items || []);
      setTotal(res.data.total || 0);
    } catch (err: any) {
      message.error(err.message || '获取项目列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchProjects(); }, []);

  const handleCreate = async (values: { name: string; description?: string }) => {
    setCreating(true);
    try {
      await api.post('/api/v1/projects', values);
      message.success('项目创建成功');
      setModalOpen(false);
      form.resetFields();
      fetchProjects();
    } catch (err: any) {
      message.error(err.message || '创建失败');
    } finally {
      setCreating(false);
    }
  };

  const columns = [
    { title: '项目名称', dataIndex: 'name', key: 'name', render: (text: string) => <span style={{ fontWeight: 500, color: '#F1F5F9' }}>{text}</span> },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', render: (t: string) => new Date(t).toLocaleDateString('zh-CN') },
    { title: '最近编辑', dataIndex: 'updated_at', key: 'updated_at', render: (t: string) => new Date(t).toLocaleDateString('zh-CN') },
    {
      title: '操作', key: 'action', width: 120,
      render: (_: any, record: Project) => (
        <Button type="primary" size="small" icon={<FolderOpenOutlined />} onClick={() => router.push(`/projects/${record.id}`)}>
          打开
        </Button>
      ),
    },
  ];

  return (
    <div style={{ padding: '32px 48px', maxWidth: 1200, margin: '0 auto' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={3} style={{ color: '#F1F5F9', margin: 0, fontFamily: "'Space Grotesk', sans-serif" }}>项目库</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>新建项目</Button>
      </div>

      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card style={{ background: '#121828', border: '1px solid #1E293B' }}>
            <Statistic title={<span style={{ color: '#94A3B8' }}>总项目数</span>} value={total} prefix={<ProjectOutlined style={{ color: '#6366F1' }} />} valueStyle={{ color: '#F1F5F9' }} />
          </Card>
        </Col>
        <Col span={8}>
          <Card style={{ background: '#121828', border: '1px solid #1E293B' }}>
            <Statistic title={<span style={{ color: '#94A3B8' }}>最近活跃</span>} value={projects.length} prefix={<ClockCircleOutlined style={{ color: '#22C55E' }} />} valueStyle={{ color: '#F1F5F9' }} />
          </Card>
        </Col>
        <Col span={8}>
          <Card style={{ background: '#121828', border: '1px solid #1E293B' }}>
            <Statistic title={<span style={{ color: '#94A3B8' }}>总资产数</span>} value={0} valueStyle={{ color: '#F1F5F9' }} />
          </Card>
        </Col>
      </Row>

      <Card style={{ background: '#121828', border: '1px solid #1E293B' }}>
        <Table columns={columns} dataSource={projects} rowKey="id" loading={loading} pagination={{ pageSize: 10 }} />
      </Card>

      <Modal title="新建项目" open={modalOpen} onCancel={() => setModalOpen(false)} footer={null} >
        <Form form={form} onFinish={handleCreate} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label={<span style={{ color: '#94A3B8' }}>项目名称</span>} rules={[{ required: true, message: '请输入项目名称' }]}>
            <Input placeholder="输入项目名称" />
          </Form.Item>
          <Form.Item name="description" label={<span style={{ color: '#94A3B8' }}>项目描述</span>}>
            <Input.TextArea rows={3} placeholder="输入项目描述（可选）" />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setModalOpen(false)}>取消</Button>
              <Button type="primary" htmlType="submit" loading={creating}>创建</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
