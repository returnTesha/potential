// app/main/dashboard/components/TurboNode.tsx
import React, { memo, ReactNode } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Cloud, Server, Database, Globe } from 'lucide-react'; // 아이콘 가져오기

// 아이콘 매핑 (노드 데이터에서 icon 이름을 문자로 받아서 실제 아이콘 컴포넌트로 변환)
const icons: Record<string, ReactNode> = {
  cloud: <Cloud size={20} />,
  server: <Server size={20} />,
  database: <Database size={20} />,
  globe: <Globe size={20} />,
};

export default memo(({ data, selected }: NodeProps) => {
  // data.icon에 맞는 아이콘을 찾고, 없으면 기본값으로 cloud 사용
  const Icon = icons[data.icon as string] || icons.cloud;

  return (
    <>
      {/* 1. 디자인된 노드 본체 */}
      <div
        style={{
          padding: '16px 20px',
          borderRadius: '12px',
          background: '#202029', // 아주 어두운 배경
          color: 'white',
          border: selected ? '2px solid #3b82f6' : '2px solid #35353d', // 선택시 파란불
          minWidth: '200px',
          display: 'flex',
          alignItems: 'center',
          gap: '12px',
          boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.5)',
          transition: 'all 0.2s',
        }}
      >
        {/* 아이콘 영역 (그라데이션 배경) */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: '40px',
            height: '40px',
            borderRadius: '8px',
            background: 'linear-gradient(135deg, #3b82f6 0%, #a855f7 100%)', // Turbo 스타일 그라데이션
          }}
        >
          {Icon}
        </div>

        {/* 텍스트 영역 */}
        <div style={{ display: 'flex', flexDirection: 'column' }}>
          <div style={{ fontSize: '16px', fontWeight: 'bold' }}>{data.title as string}</div>
          <div style={{ fontSize: '12px', color: '#9ca3af' }}>{data.subline as string}</div>
        </div>
      </div>

      {/* 2. 연결점 (핸들) - 숨겨져 있지만 기능은 함 */}
      <Handle type="target" position={Position.Left} style={{ visibility: 'hidden' }} />
      <Handle type="source" position={Position.Right} style={{ visibility: 'hidden' }} />
    </>
  );
});
