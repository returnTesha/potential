"use client"

import React from 'react';
import {
  ReactFlow,
  useNodesState,
  useEdgesState,
  Background,
  Controls,
  Position,
} from '@xyflow/react';

import '@xyflow/react/dist/style.css';
// import '@/styles/xy-theme.css'; // ⚠️ Turbo 모드에서는 기본 테마 CSS가 방해될 수 있어 주석 처리하거나 빼는 게 낫습니다.

import AnimatedSVGEdge from './animated_svg_edge';
import TurboNode from './turbo-node'; // 👈 방금 만든 노드 가져오기

// 1. 노드 타입 등록
const nodeTypes = {
  turbo: TurboNode,
};

const edgeTypes = {
  animatedSvg: AnimatedSVGEdge,
};

// 2. 노드 데이터 (Turbo 스타일로 데이터 구조 변경: title, subline, icon)
const initialNodes = [
  // [DMS]
  {
    id: 'dms',
    type: 'turbo', // 👈 타입을 'turbo'로 지정
    position: { x: 0, y: 0 },
    data: {
      icon: 'globe',
      title: 'DMS',
      subline: 'Data Management System'
    },
  },

  // [Server]
  {
    id: 'server',
    type: 'turbo',
    position: { x: 350, y: 0 },
    data: {
      icon: 'server',
      title: 'IDC46 Server',
      subline: '222.122.47.46'
    },
  },

  // [DBs]
  {
    id: 'oracle19',
    type: 'turbo',
    position: { x: 750, y: -180 },
    data: { icon: 'database', title: 'Oracle 19c', subline: 'Main DB Cluster' },
  },
  {
    id: 'oracle11',
    type: 'turbo',
    position: { x: 750, y: -60 },
    data: { icon: 'database', title: 'Oracle 11g', subline: 'Legacy System' },
  },
  {
    id: 'postgres',
    type: 'turbo',
    position: { x: 750, y: 60 },
    data: { icon: 'cloud', title: 'PostgreSQL', subline: 'v16.3 / Analytics' },
  },
  {
    id: 'mariadb',
    type: 'turbo',
    position: { x: 750, y: 180 },
    data: { icon: 'database', title: 'MariaDB', subline: 'Web Service DB' },
  },
];

const initialEdges = [
  { id: 'e1', source: 'dms', target: 'server', type: 'animatedSvg', animated: true },
  { id: 'e2', source: 'server', target: 'oracle19', type: 'animatedSvg', animated: true },
  { id: 'e3', source: 'server', target: 'oracle11', type: 'animatedSvg', animated: true },
  { id: 'e4', source: 'server', target: 'postgres', type: 'animatedSvg', animated: true },
  { id: 'e5', source: 'server', target: 'mariadb', type: 'animatedSvg', animated: true },
];

const EdgesFlow = () => {
  const [nodes, , onNodesChange] = useNodesState(initialNodes);
  const [edges, , onEdgesChange] = useEdgesState(initialEdges);

  return (
    // 3. 배경색을 짙은 남색(#1A192B)으로 설정하여 Turbo 느낌 내기
    <div style={{ width: '100%', height: '600px', background: '#1A192B', borderRadius: '8px' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes} // 👈 등록 필수
        edgeTypes={edgeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        fitView
      >
        {/* 배경 패턴 색상을 어두운 테마에 맞춰 변경 */}
        <Background color="#444" gap={20} />
        <Controls style={{ fill: 'white' }} /> {/* 컨트롤 버튼도 보이게 */}
      </ReactFlow>
    </div>
  );
};

export default EdgesFlow;
