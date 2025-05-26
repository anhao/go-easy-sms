package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"sync"
	"time"
)

var (
	// 默认客户端实例
	defaultClient *Client
	// 单例锁
	once sync.Once
)

// Client 是HTTP客户端的封装
type Client struct {
	client    *http.Client
	transport *http.Transport
	timeout   time.Duration
}

// ClientOption 是客户端配置选项
type ClientOption func(*Client)

// NewClient 创建一个新的HTTP客户端
func NewClient(options ...ClientOption) *Client {

	c := &Client{
		timeout: 5 * time.Second, // 默认超时时间
	}

	// 应用选项
	for _, option := range options {
		option(c)
	}

	// 创建HTTP客户端
	c.client = &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   c.timeout,
	}

	return c
}

// GetDefaultClient 获取默认的HTTP客户端实例（单例模式）
func GetDefaultClient() *Client {
	once.Do(func() {
		defaultClient = NewClient()
	})
	return defaultClient
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithMaxIdleConns 设置最大空闲连接数
func WithMaxIdleConns(n int) ClientOption {
	return func(c *Client) {
		c.transport.MaxIdleConns = n
	}
}

// WithMaxIdleConnsPerHost 设置每个主机的最大空闲连接数
func WithMaxIdleConnsPerHost(n int) ClientOption {
	return func(c *Client) {
		c.transport.MaxIdleConnsPerHost = n
	}
}

// Get 发送GET请求
func (c *Client) Get(ctx context.Context, urlStr string, query map[string]string, headers map[string]string) ([]byte, error) {
	req, err := c.createRequest(ctx, "GET", urlStr, nil, query, headers)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// Post 发送POST请求（表单数据）
func (c *Client) Post(ctx context.Context, urlStr string, formData map[string]string, headers map[string]string) ([]byte, error) {
	// 构建表单数据
	form := neturl.Values{}
	for k, v := range formData {
		form.Add(k, v)
	}

	// 设置Content-Type
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, exists := headers["Content-Type"]; !exists {
		headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	req, err := c.createRequest(ctx, "POST", urlStr, strings.NewReader(form.Encode()), nil, headers)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// PostJSON 发送POST请求（JSON数据）
func (c *Client) PostJSON(ctx context.Context, urlStr string, data any, headers map[string]string) ([]byte, error) {
	// 将数据转换为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// 设置Content-Type
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, exists := headers["Content-Type"]; !exists {
		headers["Content-Type"] = "application/json"
	}

	req, err := c.createRequest(ctx, "POST", urlStr, bytes.NewReader(jsonData), nil, headers)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// 创建HTTP请求
func (c *Client) createRequest(ctx context.Context, method, urlStr string, body io.Reader, query map[string]string, headers map[string]string) (*http.Request, error) {
	// 如果有查询参数，添加到URL
	if len(query) > 0 {
		queryValues := neturl.Values{}
		for k, v := range query {
			queryValues.Add(k, v)
		}
		if strings.Contains(urlStr, "?") {
			urlStr = fmt.Sprintf("%s&%s", urlStr, queryValues.Encode())
		} else {
			urlStr = fmt.Sprintf("%s?%s", urlStr, queryValues.Encode())
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, err
	}

	// 添加请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

// 执行HTTP请求
func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// 记录关闭错误，但不影响主流程
			fmt.Printf("Error closing response body: %v\n", closeErr)
		}
	}()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) Client() *http.Client {
	return c.client
}

// ParseJSONResponse 解析JSON响应
func ParseJSONResponse(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
