import React, { useState } from 'react';
import axios from 'axios';

import { SearchOutlined } from '@ant-design/icons';
import { Layout, Flex, Row, Col, Input, Button, Table, Tabs } from 'antd';
import type { TabsProps } from 'antd';


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

interface RowForChannel {

}

interface RowForChannel {
    channel_id: string;
    channel_name: string;
    subscriber_count: number;
    thumbnail_url: string;
    url: string;
    like_rate: number;
    view_count: number;
    avg_daily_views: number;
    days_ago: number;
    lasted_upload_at: Date; // time.Timeは文字列に変換する必要があります
    channel_created_at: Date; // time.Timeは文字列に変換する必要があります
    created_at: Date; // time.Timeは文字列に変換する必要があります
}

interface RowForComment {
    videoId: string;
    like: number;
    comments: number;
    name: string;
    nameId: string;
    comment: string;
    favo: number;
    replay: number;
    createdAt: Date;
}


// json rows to columns
const columnsForChannels = [
    {
        width: "5rem",
        title: 'channel_id',
        dataIndex: 'channel_id',
        key: 'channel_id',
        render: (text: string) => <a href={`https://www.youtube.com/channel/${text}`} target="_blank" rel="noreferrer">{text}</a>,
    },
    {
        width: "5rem",
        title: 'channel_name',
        dataIndex: 'channel_name',
        key: 'channel_name',
    },
    {
        width: "5rem",
        title: 'subscriber_count',
        dataIndex: 'subscriber_count',
        key: 'subscriber_count',
    },
    {
        width: "5rem",
        title: 'thumbnail_url',
        dataIndex: 'thumbnail_url',
        key: 'thumbnail_url',
        render: (text: string) => <img src={text} alt="thumbnail" style={{ width: "100px" }} />,
    },
    {
        width: "5rem",
        title: 'url',
        dataIndex: 'url',
        key: 'url',
        render: (text: string) => <a href={text} target="_blank" rel="noreferrer">{text}</a>,
    },
    {
        width: "5rem",
        title: 'like_rate',
        dataIndex: 'like_rate',
        key: 'like_rate',
    },
    {
        width: "5rem",
        title: 'view_count',
        dataIndex: 'view_count',
        key: 'view_count',
    },
    {
        width: "5rem",
        title: 'avg_daily_views',
        dataIndex: 'avg_daily_views',
        key: 'avg_daily_views',
    },
    {
        width: "5rem",
        title: 'days_ago',
        dataIndex: 'days_ago',
        key: 'days_ago',
    },
    {
        width: "5rem",
        title: 'lasted_upload_at',
        dataIndex: 'lasted_upload_at',
        key: 'lasted_upload_at',
        render: (text: string) => text.slice(0, 16),
    },
    {
        width: "5rem",
        title: 'channel_created_at',
        dataIndex: 'channel_created_at',
        key: 'channel_created_at',
        render: (text: string) => text.slice(0, 16),
    }
];

const columnsForComments = [
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
    const [rowsForChannel, setRowsForChannel] = useState<RowForChannel[]>([]);
    const [rowsForComment, setRowsForComment] = useState<RowForComment[]>([]);

    const request_channels = () => {
        const url = import.meta.env.VITE_API_URL + '/api/youtube/channels?q=' + query + '&subscribers_n=' + 10 + '&days=' + 365;
        console.log("Requesting to channels: " + url);

        axios.get(url)
            .then(res => {
                const link = document.createElement('a');
                link.href = res.data.link;
                const filename = 'youtube-data-channel-' + query + '.csv';
                link.setAttribute('download', filename);
                document.body.appendChild(link);
                link.click();

                console.log(res.data.rows);
                

                setRowsForChannel(res.data.rows);
            })
            .catch(error => {
                console.error("Error:", error);
            });
    };

    const request_comments = () => {
        const url = import.meta.env.VITE_API_URL + '/api/youtube/comments?q=' + query
        console.log("Requesting to comment: " + url);

        axios.get(url)
            .then(res => {
                const link = document.createElement('a');
                link.href = res.data.link;
                const filename = 'youtube-data-comment-' + query + '.csv';
                link.setAttribute('download', filename);
                document.body.appendChild(link);
                link.click();

                setRowsForComment(res.data.rows);
            })
            .catch(error => {
                console.error("Error:", error);
            });
    };

    const handler = (e: React.ChangeEvent<HTMLInputElement>) => {
        setQuery(e.target.value);
    };


    const tabs: TabsProps['items'] = [
        // {
        //     key: '1',
        //     label: 'Channel',
        //     children: (
        //         <>
        //             <h2>チャンネルデータ</h2>
        //             <Row gutter={16} style={{ marginBottom: "2rem" }}>

        //                 <Col span={12}>
        //                     <Input type='text' size='large' value={query} onChange={handler} placeholder="チャンネル検索ワード" />
        //                 </Col>
        //                 <Col>
        //                     <Button type='primary' size='large' icon={<SearchOutlined />} onClick={() => request_channels()}>Search</Button>
        //                 </Col>
        //             </Row>

        //             <Row>
        //                 <Col>
        //                     <Table dataSource={rowsForChannel} columns={columnsForChannels} />
        //                 </Col>
        //             </Row>
        //         </>
        //     ),
        // },
        {
            key: '2',
            label: 'Comment',
            children: (
                <>
                    <h2>コメントデータ</h2>
                    <Row gutter={16} style={{ marginBottom: "2rem" }}>

                        <Col span={12}>
                            <Input type='text' size='large' value={query} onChange={handler} placeholder="動画検索ワード" />
                        </Col>
                        <Col>
                            <Button type='primary' size='large' icon={<SearchOutlined />} onClick={() => request_comments()}>Search</Button>
                        </Col>
                    </Row>

                    <Row>
                        <Col>
                            <Table dataSource={rowsForComment} columns={columnsForComments} />
                        </Col>
                    </Row>
                </>
            ),
        }
    ];

    return (
        <Flex gap="middle" wrap="wrap">

            <Layout style={layoutStyle}>
                <Header style={headerStyle}></Header>
                <Content style={contentStyle}>
                    <Tabs defaultActiveKey="1" items={tabs} />
                </Content>
                <Footer style={footerStyle}>Youtube search query © {new Date().getFullYear()} Created by {import.meta.env.VITE_USERNAME}</Footer>
            </Layout>

        </Flex>
    );
}

export default App;