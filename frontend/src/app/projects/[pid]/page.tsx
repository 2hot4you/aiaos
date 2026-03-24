'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Button, Card, Collapse, Typography, Space, Modal, Form, Input, InputNumber, message, Row, Col, Statistic, Empty, Popconfirm } from 'antd';
import { PlusOutlined, ArrowLeftOutlined, PlayCircleOutlined, DeleteOutlined } from '@ant-design/icons';
import api from '@/lib/api';
import type { Season, Episode, Project } from '@/types';

const { Title, Text } = Typography;

export default function ProjectDetailPage() {
  const params = useParams();
  const router = useRouter();
  const pid = params.pid as string;

  const [project, setProject] = useState<Project | null>(null);
  const [seasons, setSeasons] = useState<Season[]>([]);
  const [episodes, setEpisodes] = useState<Record<string, Episode[]>>({});
  const [loading, setLoading] = useState(true);
  const [seasonModalOpen, setSeasonModalOpen] = useState(false);
  const [episodeModalOpen, setEpisodeModalOpen] = useState(false);
  const [currentSeasonId, setCurrentSeasonId] = useState<string>('');
  const [seasonForm] = Form.useForm();
  const [episodeForm] = Form.useForm();

  const fetchProject = async () => {
    try {
      const res: any = await api.get(`/api/v1/projects/${pid}`);
      setProject(res.data);
    } catch {}
  };

  const fetchSeasons = async () => {
    setLoading(true);
    try {
      const res: any = await api.get(`/api/v1/projects/${pid}/seasons`);
      const seasonList = res.data.items || res.data || [];
      setSeasons(seasonList);
      for (const season of seasonList) {
        fetchEpisodes(season.id);
      }
    } catch (err: any) {
      message.error(err.message || '获取季列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchEpisodes = async (seasonId: string) => {
    try {
      const res: any = await api.get(`/api/v1/seasons/${seasonId}/episodes`);
      setEpisodes((prev) => ({ ...prev, [seasonId]: res.data.items || res.data || [] }));
    } catch {}
  };

  useEffect(() => { fetchProject(); fetchSeasons(); }, [pid]);

  const handleCreateSeason = async (values: { title: string; season_number: number }) => {
    try {
      await api.post(`/api/v1/projects/${pid}/seasons`, values);
      message.success('季创建成功');
      setSeasonModalOpen(false);
      seasonForm.resetFields();
      fetchSeasons();
    } catch (err: any) { message.error(err.message || '创建失败'); }
  };

  const handleCreateEpisode = async (values: { title: string; episode_number: number }) => {
    try {
      await api.post(`/api/v1/seasons/${currentSeasonId}/episodes`, values);
      message.success('集创建成功');
      setEpisodeModalOpen(false);
      episodeForm.resetFields();
      fetchEpisodes(currentSeasonId);
    } catch (err: any) { message.error(err.message || '创建失败'); }
  };

  const totalEpisodes = Object.values(episodes).reduce((sum, eps) => sum + eps.length, 0);

  const collapseItems = seasons.map((season) => ({
    key: season.id,
    label: (
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%' }}>
        <Text style={{ color: '#F1F5F9', fontWeight: 500 }}>第{season.season_number}季：{season.title}</Text>
        <Text style={{ color: '#64748B', fontSize: 12 }}>{(episodes[season.id] || []).length} 集</Text>
      </div>
    ),
    children: (
      <div>
        {(episodes[season.id] || []).length === 0 ? (
          <Empty description={<Text style={{ color: '#64748B' }}>暂无集数</Text>} image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {(episodes[season.id] || []).map((ep) => (
              <Card key={ep.id} size="small" style={{ background: '#1E293B', border: '1px solid #334155', cursor: 'pointer' }}
                hoverable
                onClick={() => router.push(`/projects/${pid}/seasons/${season.id}/episodes/${ep.id}/script`)}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Space>
                    <PlayCircleOutlined style={{ color: '#6366F1' }} />
                    <Text style={{ color: '#F1F5F9' }}>第{ep.episode_number}集：{ep.title}</Text>
                  </Space>
                  <Button type="primary" size="small" ghost>进入工作台</Button>
                </div>
              </Card>
            ))}
          </div>
        )}
        <Button type="dashed" icon={<PlusOutlined />} block style={{ marginTop: 12, borderColor: '#334155', color: '#94A3B8' }}
          onClick={() => { setCurrentSeasonId(season.id); episodeForm.setFieldsValue({ episode_number: (episodes[season.id] || []).length + 1 }); setEpisodeModalOpen(true); }}>
          添加集
        </Button>
      </div>
    ),
  }));

  return (
    <div style={{ padding: '32px 48px', maxWidth: 1200, margin: '0 auto' }}>
      <Space style={{ marginBottom: 24 }}>
        <Button icon={<ArrowLeftOutlined />} type="text" style={{ color: '#94A3B8' }} onClick={() => router.push('/projects')}>返回项目库</Button>
      </Space>

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={3} style={{ color: '#F1F5F9', margin: 0, fontFamily: "'Space Grotesk', sans-serif" }}>{project?.name || '加载中...'}</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { seasonForm.setFieldsValue({ season_number: seasons.length + 1 }); setSeasonModalOpen(true); }}>新建季</Button>
      </div>

      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}><Card style={{ background: '#121828', border: '1px solid #1E293B' }}><Statistic title={<span style={{ color: '#94A3B8' }}>总季数</span>} value={seasons.length} valueStyle={{ color: '#F1F5F9' }} /></Card></Col>
        <Col span={6}><Card style={{ background: '#121828', border: '1px solid #1E293B' }}><Statistic title={<span style={{ color: '#94A3B8' }}>总集数</span>} value={totalEpisodes} valueStyle={{ color: '#F1F5F9' }} /></Card></Col>
        <Col span={6}><Card style={{ background: '#121828', border: '1px solid #1E293B' }}><Statistic title={<span style={{ color: '#94A3B8' }}>角色数</span>} value={0} valueStyle={{ color: '#F1F5F9' }} /></Card></Col>
        <Col span={6}><Card style={{ background: '#121828', border: '1px solid #1E293B' }}><Statistic title={<span style={{ color: '#94A3B8' }}>场景数</span>} value={0} valueStyle={{ color: '#F1F5F9' }} /></Card></Col>
      </Row>

      <Card style={{ background: '#121828', border: '1px solid #1E293B' }}>
        {seasons.length === 0 ? (
          <Empty description={<Text style={{ color: '#64748B' }}>还没有季，点击上方"新建季"开始</Text>} image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <Collapse items={collapseItems} defaultActiveKey={seasons.map(s => s.id)} style={{ background: 'transparent', border: 'none' }}
            expandIconPosition="start" />
        )}
      </Card>

      <Modal title="新建季" open={seasonModalOpen} onCancel={() => setSeasonModalOpen(false)} footer={null}>
        <Form form={seasonForm} onFinish={handleCreateSeason} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="season_number" label="季号" rules={[{ required: true }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="title" label="季标题" rules={[{ required: true, message: '请输入标题' }]}><Input placeholder="第一季" /></Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}><Space><Button onClick={() => setSeasonModalOpen(false)}>取消</Button><Button type="primary" htmlType="submit">创建</Button></Space></Form.Item>
        </Form>
      </Modal>

      <Modal title="新建集" open={episodeModalOpen} onCancel={() => setEpisodeModalOpen(false)} footer={null}>
        <Form form={episodeForm} onFinish={handleCreateEpisode} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="episode_number" label="集号" rules={[{ required: true }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="title" label="集标题" rules={[{ required: true, message: '请输入标题' }]}><Input placeholder="第一集：开端" /></Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}><Space><Button onClick={() => setEpisodeModalOpen(false)}>取消</Button><Button type="primary" htmlType="submit">创建</Button></Space></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
