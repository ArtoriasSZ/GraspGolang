# Grasp HTTP
## 整体框架
### 服务端注册路由函数与启动监听
```asciidoc
func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	http.ListenAndServe(":8091", nil)
}
```
这里做了两件事情
- 调用 http.HandleFunc 方法，注册了对应于请求路径 /ping 的 handler 函数
- 调用 http.ListenAndServe，启动了一个端口为 8091 的 http 服务
细节我们后面继续探索
### 客户端发送请求
```asciidoc
func main() {
	reqBody, _ := json.Marshal(map[string]string{"key1": "val1", "key2": "val2"})

	resp, _ := http.Post(":8091", "application/json", bytes.NewReader(reqBody))
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("resp: %s", respBody)
}
```
- 序列化一个请求体，并发送一个 POST 请求到服务端。
- 用resp接受返回的响应体，并进行处理。
## 细节
- 服务端	net/http/server.go
- 客户端——主流程	net/http/client.go
- 客户端——构造请求	net/http/request.go
- 客户端——网络交互	net/http/transport.go

### 服务端
#### 核心数据结构
##### Handler
```asciidoc
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```
- Handler 是路由处理函数 用于响应 HTTP 请求。
- [Handler.ServeHTTP] 应该向 [ResponseWriter] 写入响应头和数据，然后返回。
- 返回值表示请求处理完成；在 [ServeHTTP] 调用完成后或并发进行时，不应该再使用 [ResponseWriter] 或读取 [Request.Body]。
- 根据 HTTP 客户端软件、HTTP 协议版本以及客户端和 Go 服务器之间的任何中介，可能无法在向 [ResponseWriter] 写入数据后读取 [Request.Body]。
- 谨慎的处理程序应该先读取 [Request.Body]，然后再回复。
- 除了读取请求体外，处理程序不应修改提供的 Request。
- 如果 ServeHTTP 发生宕机，服务器（调用 ServeHTTP 的调用方）假定宕机的影响仅限于活动请求。
- 它会恢复宕机，将堆栈跟踪记录到服务器错误日志，并根据 HTTP 协议关闭网络连接或发送 HTTP/2 RST_STREAM。
- 要中断处理程序，以便客户端看到中断的响应但服务器不记录错误，请使用值 [ErrAbortHandler] 抛出宕机。

##### ResponseWriter
```asciidoc
type ResponseWriter interface {
	Header() Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
	}
```
- ResponseWriter 接口是用于构建 HTTP 响应的主要接口，它定义了以下方法：
- Header() Header：返回一个 Header 类型的对象，用于设置 HTTP 响应的头部信息。头部信息包括诸如 Content-Type、Content-Length 等与响应相关的元数据。
- Write([]byte) (int, error)：将字节数据写入 HTTP 响应体，并返回写入的字节数和可能的错误。
- WriteHeader(statusCode int)：发送 HTTP 响应的状态码。如果在调用 WriteHeader 之前没有调用过，则 Write 方法会在写入数据之前自动调用 WriteHeader(http.StatusOK)。
##### Flusher
```asciidoc
type Flusher interface {
	Flush()
```
- Flusher接口由允许HTTP处理程序将缓冲的数据刷新到客户端的ResponseWriters实现。
- 默认的HTTP/1.x和HTTP/2 [ResponseWriter]实现支持[Flusher]，但是ResponseWriter包装器可能不支持。处理程序应始终在运行时测试此功能。
- 请注意，即使对于支持Flush的ResponseWriters，如果客户端通过HTTP代理连接，直到响应完成，缓冲的数据可能无法到达客户端。
- Flusher 接口是一个可选接口，由支持将缓冲数据刷新到客户端的 ResponseWriter 实现。主要用于在处理 HTTP 请求时，需要将一部分数据推送给客户端，而不是等待整个响应完成后才发送。这在处理大文件下载或者实时数据传输等场景中很有用。
###### 示例代码
```asciidoc
http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
        // 设置响应头，这个很重要
		for i := 0; i <= 10; i++ {
			w.Write([]byte(fmt.Sprintf("pong %d \n", i)))
			w.(http.Flusher).Flush()
			time.Sleep(1 * time.Second)
		}
	})
```
##### Hijacker
```asciidoc
type Hijacker interface {
	Hijack() (net.Conn, *bufio.ReadWriter, error)
}
```
- Hijacker接口由允许HTTP处理程序接管连接的ResponseWriters实现。
- 默认情况下，用于HTTP/1.x连接的[ResponseWriter]支持Hijacker，但是意图上不支持HTTP/2连接。
- ResponseWriter包装器也可能不支持Hijacker。处理程序应始终在运行时测试此功能。
- Hijack 方法允许调用者接管连接，这意味着 HTTP 服务器将不再处理该连接，而完全由调用者负责管理和关闭连接。
- 返回的 net.Conn 表示连接，而 bufio.ReadWriter 则是一个用于读写的缓冲区，其中可能包含来自客户端的未处理的缓冲数据。
- 在调用 Hijack 后，原始请求的 Request.Body 不应再被使用，但原始请求的上下文 (Context) 仍然有效，直到请求的 ServeHTTP 方法返回。
- 这种机制通常用于特殊的情况，例如在 WebSocket 或其他非标准协议中，需要更底层的控制。通过调用 Hijack 方法，你可以取得连接的完全控制权，然后按照自己的需要进行处理。
##### CloseNotifier
```asciidoc
type CloseNotifier interface {
	CloseNotify() <-chan bool
}
```
- CloseNotifier 接口由 ResponseWriters 实现，它允许检测基础连接何时消失。如果客户端在响应准备就绪之前断开连接，则此机制可用于取消服务器上的长时间操作。
- 已弃用：CloseNotifier 接口早于 Go 的上下文包。新代码应改用 [Request.Context]。
#####
```asciidoc
type conn struct {
	server *Server
	cancelCtx context.CancelFunc
	rwc net.Conn
	remoteAddr string
	tlsState *tls.ConnectionState
	werr error
	r *connReader
	bufr *bufio.Reader
	bufw *bufio.Writer
	lastMethod string
	curReq atomic.Pointer[response] // (which has a Request in it)
	curState atomic.Uint64 // packed (unixtime<<8|uint8(ConnState))
	mu sync.Mutex
	hijackedv bool
}
```
- conn 代表了一个HTTP连接的服务器端
- server 是当前连接所在的服务器
- cancelCtx 取消连接级别的上下文
- rwc 是底层网络连接。它从未被其他类型包装过，是返回给CloseNotifier调用方的值。通常为*net.TCPConn或*tls.Conn。
- remoteAddr 是rwc.RemoteAddr().String()。它不会在Listener的Accept协程内同步填充，因为某些实现会阻塞。它会在(*conn).serve协程内立即填充。这是Handler的(*Request).RemoteAddr的值。
- tlsState 是在使用TLS时的TLS连接状态。nil表示没有TLS。
- werr 设置为第一个写入到rwc的错误。它通过checkConnErrorWriter{w}设置，其中bufw是写入器。
- r 是bufr的读取源。它是rwc的包装器，提供io.LimitedReader风格的限制（在读取请求头时）和支持CloseNotifier的功能。见*connReader文档。
- bufr从r读取。
- bufw写入到checkConnErrorWriter{c}，该函数在出现错误时会填充werr。
- lastMethod是最近一个请求的方法。如果有的话，就在此连接上。
- mu 保护hijackedv
- hijackedv表示此连接是否已被具有Hijacker接口的Handler劫持。

#### response
```asciidoc
type response struct {
	conn             *conn
	req              *Request 
	reqBody          io.ReadCloser
	cancelCtx        context.CancelFunc 
	wroteHeader      bool              
	wroteContinue    bool               
	wants10KeepAlive bool               
	wantsClose       bool               
	canWriteContinue atomic.Bool
	writeContinueMu  sync.Mutex

	w  *bufio.Writer // buffers output in chunks to chunkWriter
	cw chunkWriter

	
	handlerHeader Header
	calledHeader  bool 

	written       int64 
	contentLength int64 
	status        int   

	closeAfterReply bool

	fullDuplex bool

	requestBodyLimitHit bool

	trailers []string

	handlerDone atomic.Bool
	
	dateBuf   [len(TimeFormat)]byte
	clenBuf   [10]byte
	statusBuf [3]byte

	closeNotifyCh  chan bool
	didCloseNotify atomic.Bool 
}
```
- response 表示 HTTP 响应的服务端部分。
- conn 字段是底层网络连接。
- req 是用于此响应的请求。
- reqBody 是请求体。
- 当 ServeHTTP 退出时，cancelCtx 取消上下文。
- wroteHeader 表示是否已写入非 1xx 头部。
- wroteContinue 表示是否已写入“100 Continue”响应。
- wants10KeepAlive 表示连接是否希望在 HTTP/1.0 中保持活动状态。
- wantsClose 表示 HTTP 请求是否具有连接“关闭”。
- canWriteContinue 是一个原子布尔值，指定是否可以将“100 Continue”头写入连接。
- 在编写头部时必须保持 writeContinueMu 。
- w 在 chunkWriter 中按块缓冲输出。
- cw 是 chunkWriter。

#### Request
```asciidoc
type Request struct {
	Method string
	URL *url.URL
	Proto      string // "HTTP/1.0"
	ProtoMajor int    // 1
	ProtoMinor int    // 0

	Header Header

	Body io.ReadCloser

	GetBody func() (io.ReadCloser, error)

	ContentLength int64

	TransferEncoding []string

	Close bool

	Host string

	Form url.Values

	PostForm url.Values

	MultipartForm *multipart.Form

	Trailer Header

	RemoteAddr string

	RequestURI string

	TLS *tls.ConnectionState

	Cancel <-chan struct{}

	Response *Response

	ctx context.Context

	pat         *pattern          // the pattern that matched
	matches     []string          // values for the matching wildcards in pat
	otherValues map[string]string // for calls to SetPathValue that don't match a wildcard
}
```
- Request 表示服务器接收到的 HTTP 请求或客户端将要发送的请求。
- Method 指定 HTTP 方法（GET、POST、PUT 等）。
- Proto ProtoMajor ProtoMinor用于传入服务器请求的协议版本。

#### ServeMux










