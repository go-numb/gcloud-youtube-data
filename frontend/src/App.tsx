import React, { useState } from 'react';
import axios from 'axios';

import { SearchOutlined } from '@ant-design/icons';
import { Layout, Flex, Row, Col, Input, Button, Table } from 'antd';

const { Header, Footer, Content } = Layout;

const headerStyle: React.CSSProperties = {
    textAlign: 'center',
    height: 64,
    paddingInline: 48,
    lineHeight: '64px',
    backgroundColor: '#fff',
};

const layoutStyle = {
    backgroundColor: '#fff',
};

const contentStyle: React.CSSProperties = {
    maxWidth: "1280px",
    margin: "0 auto",
    padding: "2rem",
    textAlign: "center",
};

const footerStyle: React.CSSProperties = {
    textAlign: 'center',
    color: '#ddd',
    backgroundColor: '#fff',
};

interface Row {
    videoId: string;
    like: number;
    comments: number;
    name: string;
    nameId: string;
    comment: string;
    favo: number;
    replay: number;
    createdAt: string;
}

// json rows to columns
const columns = [
    {
        width: "5rem",
        title: 'video_id',
        dataIndex: 'video_id',
        key: 'video_id',
        render: (text: string) => <a href={`https://www.youtube.com/watch?v=${text}`} target="_blank" rel="noreferrer">{text}</a>,
    },
    {
        width: "5rem",
        title: 'like',
        dataIndex: 'like',
        key: 'like',
    },
    {
        width: "5rem",
        title: 'comments',
        dataIndex: 'comments',
        key: 'comments',
    },
    {
        width: "5rem",
        title: 'name',
        dataIndex: 'name',
        key: 'name',
    },
    {
        width: "5rem",
        title: 'name_id',
        dataIndex: 'name_id',
        key: 'name_id',
        render: (text: string) => <a href={`https://www.youtube.com/channel/${text}`} target="_blank" rel="noreferrer">{text}</a>,
    },
    {
        width: "15rem",
        title: 'comment',
        dataIndex: 'comment',
        key: 'comment',
        render: (text: string) => text.slice(0, 140) + '...',
    },
    {
        width: "5rem",
        title: 'favo',
        dataIndex: 'favo',
        key: 'favo',
    },
    {
        width: "5rem",
        title: 'replay',
        dataIndex: 'replay',
        key: 'replay',
    },
    {
        width: "5rem",
        title: 'created_at',
        dataIndex: 'created_at',
        key: 'created_at',
        render: (text: string) => text.slice(0, 16),
    },
];




function App() {
    const [query, setQuery] = useState("");

    // csv download to rows
    const [rows, setRows] = useState<Row[]>([]);

    const request = () => {
        console.log("Requesting: " + query);

        axios.get(import.meta.env.VITE_API_URL + '/api/youtube?q=' + query)
            .then(res => {
                const link = document.createElement('a');
                link.href = res.data.link;
                const filename = 'youtube-' + query + '.csv';
                link.setAttribute('download', filename);
                document.body.appendChild(link);
                link.click();

                setRows(res.data.rows);
            })
            .catch(error => {
                console.error("Error:", error);
            });
    };

    const handler = (e: React.ChangeEvent<HTMLInputElement>) => {
        setQuery(e.target.value);
    };

    return (
        <Flex gap="middle" wrap="wrap">
            
            <Layout style={layoutStyle}>
                <Header style={headerStyle}></Header>
                <Content style={contentStyle}>

                    <Row gutter={16} style={{ marginBottom: "2rem" }}>
                        <Col span={12}>
                            <Input type='text' size='large' value={query} onChange={handler} placeholder="検索ワード" />
                        </Col>
                        <Col>
                            <Button type='primary' size='large' icon={<SearchOutlined />} onClick={() => request()}>Search</Button>
                        </Col>
                    </Row>

                    <Row>
                        <Col>
                            <Table dataSource={rows} columns={columns} />;
                        </Col>
                    </Row>

                </Content>
                <Footer style={footerStyle}>Youtube search query © {new Date().getFullYear()} Created by { import.meta.env.VITE_USERNAME }</Footer>
            </Layout>

        </Flex>
    );
}

export default App;