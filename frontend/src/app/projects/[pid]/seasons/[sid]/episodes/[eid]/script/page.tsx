'use client';

import { useState } from 'react';
import { Card, Select, Radio, Input, Button, Typography, Space, Divider } from 'antd';
import { ThunderboltOutlined } from '@ant-design/icons';

const { Title, Text } = Typography;
const { TextArea } = Input;

export default function ScriptPage() {
  const [script, setScript] = useState('');
  const [mode, setMode] = useState('novel');

  return (
    <div style={{ display: 'flex', height: '100%', gap: 0 }}>
      {/* 中间 - 项目配置 */}
      <div style={{ width: 360, borderRight: '1px solid #1E293B', padding: 24, overflowY: 'auto', background: '#0A0E1A' }}>
        <Title level={5} style={{ color: '#F1F5F9', fontFamily: "'Space Grotesk', sans-serif", marginBottom: 20 }}>
          项目配置
        </Title>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div>
            <Text style={{ color: '#94A3B8', fontSize: 12, marginBottom: 6, display: 'block' }}>创作模式</Text>
            <Radio.Group value={mode} onChange={(e) => setMode(e.target.value)} style={{ width: '100%' }}>
              <Space direction="vertical" style={{ width: '100%' }}>
                <Radio value="novel" style={{ color: '#F1F5F9' }}>
                  <div>
                    <div style={{ fontWeight: 500 }}>小说生成分镜</div>
                    <Text style={{ color: '#64748B', fontSize: 11 }}>适合粘贴小说/章节/大纲</Text>
                  </div>
                </Radio>
                <Radio value="storyboard" style={{ color: '#F1F5F9' }}>
                  <div>
                    <div style={{ fontWeight: 500 }}>分镜生成分镜</div>
                    <Text style={{ color: '#64748B', fontSize: 11 }}>按已有分镜逐条生成</Text>
                  </div>
                </Radio>
              </Space>
            </Radio.Group>
          </div>

          <div>
            <Text style={{ color: '#94A3B8', fontSize: 12, marginBottom: 6, display: 'block' }}>输出语言</Text>
            <Select defaultValue="zh" style={{ width: '100%' }} options={[{ value: 'zh', label: '中文' }, { value: 'en', label: 'English' }]} />
          </div>

          <div>
            <Text style={{ color: '#94A3B8', fontSize: 12, marginBottom: 6, display: 'block' }}>画面比例</Text>
            <Radio.Group defaultValue="landscape" buttonStyle="solid" style={{ width: '100%' }}>
              <Radio.Button value="landscape" style={{ width: '50%', textAlign: 'center' }}>横屏 16:9</Radio.Button>
              <Radio.Button value="portrait" style={{ width: '50%', textAlign: 'center' }}>竖屏 9:16</Radio.Button>
            </Radio.Group>
          </div>

          <div>
            <Text style={{ color: '#94A3B8', fontSize: 12, marginBottom: 6, display: 'block' }}>目标时长</Text>
            <Select defaultValue="60" style={{ width: '100%' }} options={[
              { value: '30', label: '30 秒' },
              { value: '60', label: '1 分钟' },
              { value: '120', label: '2 分钟' },
              { value: '300', label: '5 分钟' },
            ]} />
          </div>

          <div>
            <Text style={{ color: '#94A3B8', fontSize: 12, marginBottom: 6, display: 'block' }}>分镜生成模型</Text>
            <Select defaultValue="gpt" style={{ width: '100%' }} options={[
              { value: 'gpt', label: 'GPT' },
            ]} />
          </div>

          <div>
            <Text style={{ color: '#94A3B8', fontSize: 12, marginBottom: 6, display: 'block' }}>视觉风格</Text>
            <Select defaultValue="anime" style={{ width: '100%' }} options={[
              { value: 'anime', label: '日式动漫' },
              { value: '2d', label: '2D 动画' },
              { value: '3d', label: '3D 动画' },
              { value: 'cyberpunk', label: '赛博朋克' },
              { value: 'oil', label: '油画风格' },
              { value: 'realistic', label: '真人影视' },
            ]} />
          </div>

          <Divider style={{ borderColor: '#1E293B', margin: '8px 0' }} />

          <Button type="primary" icon={<ThunderboltOutlined />} size="large" block style={{ height: 48, fontWeight: 600, borderRadius: 8 }}>
            生成分镜剧本
          </Button>
        </div>
      </div>

      {/* 右侧 - 剧本编辑器 */}
      <div style={{ flex: 1, padding: 24, display: 'flex', flexDirection: 'column', background: '#0A0E1A' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <Title level={5} style={{ color: '#F1F5F9', margin: 0, fontFamily: "'Space Grotesk', sans-serif" }}>
            剧本编辑器
          </Title>
          <Space>
            <Button size="small" disabled>AI 改写</Button>
            <Button size="small" disabled>AI 续写</Button>
          </Space>
        </div>

        <TextArea
          value={script}
          onChange={(e) => setScript(e.target.value)}
          placeholder="在这里输入你的剧本内容...&#10;&#10;可以是完整的小说章节、剧情大纲、或已经写好的分镜描述。&#10;&#10;示例：&#10;第一幕：清晨，城市天际线。&#10;阳光透过摩天大楼的缝隙洒下金色的光芒。&#10;主角站在天台上，望着远方..."
          style={{
            flex: 1,
            background: '#121828',
            border: '1px solid #1E293B',
            color: '#F1F5F9',
            fontSize: 14,
            lineHeight: 1.8,
            borderRadius: 8,
            resize: 'none',
            fontFamily: "'DM Sans', sans-serif",
          }}
        />

        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 12 }}>
          <Text style={{ color: '#64748B', fontSize: 12 }}>{script.length} 字</Text>
          <Text style={{ color: '#64748B', fontSize: 12 }}>自动保存</Text>
        </div>
      </div>
    </div>
  );
}
