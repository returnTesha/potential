"use client"

import React from 'react';
import {
  ReactFlow,
  useNodesState,
  useEdgesState,
  Background,
  Controls,
  Position, // 👈 선을 옆으로 연결하기 위해 추가했습니다.
} from '@xyflow/react';

import '@xyflow/react/dist/style.css';
import '@/styles/xy-theme.css';

import AnimatedSVGEdge from './animated_svg_edge';

// 1. 노드 정의 (좌 -> 우 배치)
const initialNodes = [
  // [1단계] 가장 왼쪽: IDC
  {
    id: 'dms',
    type: 'input', // 시작점
    sourcePosition: Position.Right, // 오른쪽으로 선이 나감
    position: { x: 0, y: 0 },
    data: { label: 'DMS' },
  },

  // [2단계] 중간: 서버 (222.122.47.46)
  {
    id: 'server',
    sourcePosition: Position.Right, // 오른쪽으로 선이 나감
    targetPosition: Position.Left,  // 왼쪽에서 선을 받음
    position: { x: 300, y: 0 },     // X축 300 이동
    data: { label: '222.122.47.46' },
  },

  // [3단계] 가장 오른쪽: DB 4개 (Y축으로 펼침)
  {
    id: 'oracle19',
    targetPosition: Position.Left, // 왼쪽에서 선을 받음
    position: { x: 650, y: -150 }, // 위로
    data: { label: 'Oracle 19c' },
  },
  {
    id: 'oracle11',
    targetPosition: Position.Left,
    position: { x: 650, y: -50 },
    data: { label: 'Oracle 11g' },
  },
  {
    id: 'postgres',
    targetPosition: Position.Left,
    position: { x: 650, y: 50 },
    data: { label: 'Postgresql 16.3' },
  },
  {
    id: 'mariadb',
    targetPosition: Position.Left,
    position: { x: 650, y: 150 }, // 아래로
    data: { label: 'MariaDB' },
  },
];

const edgeTypes = {
  animatedSvg: AnimatedSVGEdge,
};

// 2. 엣지 연결 (IDC -> Server -> DBs)
const initialEdges = [
  // IDC -> Server
  {
    id: 'e-dms-server',
    source: 'dms',
    target: 'server',
    type: 'animatedSvg',
    animated: true
  },

  // Server -> DBs
  { id: 'e-server-ora19', source: 'server', target: 'oracle19', type: 'animatedSvg', animated: true },
  { id: 'e-server-ora11', source: 'server', target: 'oracle11', type: 'animatedSvg', animated: true },
  { id: 'e-server-pg',    source: 'server', target: 'postgres', type: 'animatedSvg', animated: true },
  { id: 'e-server-maria', source: 'server', target: 'mariadb',  type: 'animatedSvg', animated: true },
];

const EdgesFlow = () => {
  const [nodes, , onNodesChange] = useNodesState(initialNodes);
  const [edges, , onEdgesChange] = useEdgesState(initialEdges);

  return (
    // 높이(height)를 반드시 지정해야 화면에 보입니다.
    <div style={{ width: '100%', height: '500px', border: '1px solid #ddd', borderRadius: '8px' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        edgeTypes={edgeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        fitView // 시작할 때 그래프 전체가 보이도록 자동 줌
      >
        <Background />
        <Controls /> {/* 확대/축소 버튼 추가 */}
      </ReactFlow>
    </div>
  );
};

export default EdgesFlow;
