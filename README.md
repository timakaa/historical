# gRPC Hello World Example

Этот проект демонстрирует базовую реализацию gRPC клиента и сервера на языке Go.

## Содержание

- [Обзор](#обзор)
- [Структура проекта](#структура-проекта)
- [Protocol Buffers](#protocol-buffers)
- [Реализация сервера](#реализация-сервера)
- [Реализация клиента](#реализация-клиента)
- [Запуск приложения](#запуск-приложения)
- [Обработка ошибок в gRPC](#обработка-ошибок-в-grpc)
- [Дополнительные возможности](#дополнительные-возможности)

## Обзор

gRPC - это высокопроизводительный фреймворк для удаленного вызова процедур (RPC), разработанный Google. Он использует Protocol Buffers для определения интерфейса сервиса и формата сообщений, а также HTTP/2 для транспорта.

Основные преимущества gRPC:

1. **Высокая производительность** - использует HTTP/2 и бинарный протокол
2. **Строгая типизация** - контракты API определяются с помощью Protocol Buffers
3. **Поддержка многих языков** - автоматическая генерация клиентского и серверного кода
4. **Двунаправленный стриминг** - поддержка потоковой передачи данных в обоих направлениях

## Структура проекта

```
crypto_prices/
├── cmd/
│   ├── app/
│   │   └── main.go         # Основное приложение (клиент + сервер)
│   ├── client/
│   │   └── main.go         # Отдельный клиент
│   └── server/
│       └── main.go         # Отдельный сервер
├── internal/
│   ├── client/
│   │   └── client.go       # Реализация gRPC клиента
│   └── server/
│       └── server.go       # Реализация gRPC сервера
└── proto/
    ├── hello.proto         # Определение Protocol Buffers
    ├── hello.pb.go         # Сгенерированный код для сообщений
    └── hello_grpc.pb.go    # Сгенерированный код для сервиса
```

## Protocol Buffers

Protocol Buffers (protobuf) - это механизм сериализации структурированных данных, разработанный Google. В gRPC он используется для определения:

1. Структуры сообщений (messages)
2. Методов сервиса (service methods)

Наш файл `proto/hello.proto`:

```protobuf
syntax = "proto3";

package hello;

option go_package = "./proto";

// The greeting service definition
service Greeter {
  // Sends a greeting
  rpc SayHello (HelloRequest) returns (HelloReply) {}
}

// The request message containing the user's name
message HelloRequest {
  string name = 1;
}

// The response message containing the greeting
message HelloReply {
  string message = 1;
}
```

### Генерация кода

Для генерации Go-кода из proto-файла используются инструменты:

1. `protoc` - компилятор Protocol Buffers
2. `protoc-gen-go` - плагин для генерации Go-кода для сообщений
3. `protoc-gen-go-grpc` - плагин для генерации Go-кода для gRPC сервиса

Команда для генерации:

```bash
protoc --go_out=. --go-grpc_out=. proto/hello.proto
```

Эта команда создает два файла:

- `hello.pb.go` - содержит код для сообщений (HelloRequest, HelloReply)
- `hello_grpc.pb.go` - содержит код для сервиса (интерфейсы клиента и сервера)

## Реализация сервера

Сервер реализован в файле `internal/server/server.go`:

```go
package server

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "crypto_prices/proto"

	"google.golang.org/grpc"
)

// Server implements the gRPC Greeter service
type Server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements the SayHello RPC method
func (s *Server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received request from: %s", req.GetName())
	return &pb.HelloReply{Message: fmt.Sprintf("Hello, %s!", req.GetName())}, nil
}

// Start starts the gRPC server on the specified port
func Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &Server{})

	log.Printf("Server listening on port %d", port)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
```

### Ключевые компоненты сервера:

1. **Структура Server** - реализует интерфейс `GreeterServer`, определенный в сгенерированном коде

   - Встраивает `UnimplementedGreeterServer` для совместимости с будущими версиями API

2. **Метод SayHello** - реализует RPC-метод, определенный в proto-файле

   - Принимает контекст и запрос (`HelloRequest`)
   - Возвращает ответ (`HelloReply`) и ошибку

3. **Функция Start** - запускает gRPC сервер
   - Создает TCP-слушатель на указанном порту
   - Создает экземпляр gRPC сервера
   - Регистрирует реализацию сервиса
   - Запускает сервер для обработки запросов

## Реализация клиента

Клиент реализован в файле `internal/client/client.go`:

```go
package client

import (
	"context"
	"fmt"
	"time"

	pb "crypto_prices/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a gRPC client for the Greeter service
type Client struct {
	conn   *grpc.ClientConn
	client pb.GreeterClient
}

// NewClient creates a new gRPC client connected to the specified address
func NewClient(address string) (*Client, error) {
	// Set up a connection to the server with insecure credentials (no TLS)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	client := pb.NewGreeterClient(conn)
	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// SayHello sends a greeting to the server
func (c *Client) SayHello(name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := c.client.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		return "", fmt.Errorf("could not greet: %v", err)
	}

	return resp.GetMessage(), nil
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.conn.Close()
}
```

### Ключевые компоненты клиента:

1. **Структура Client** - инкапсулирует соединение и клиентский стаб

   - `conn` - соединение gRPC
   - `client` - сгенерированный клиентский стаб для вызова методов

2. **Функция NewClient** - создает новый экземпляр клиента

   - Устанавливает соединение с сервером (без TLS для простоты)
   - Создает клиентский стаб

3. **Метод SayHello** - вызывает удаленный метод на сервере

   - Создает контекст с таймаутом
   - Формирует запрос и отправляет его серверу
   - Возвращает сообщение из ответа

4. **Метод Close** - закрывает соединение с сервером

## Запуск приложения

Основное приложение (`cmd/app/main.go`) запускает и сервер, и клиент в одном процессе:

```go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"crypto_prices/internal/client"
	"crypto_prices/internal/server"
)

func main() {
	const port = 50051
	const serverAddress = "localhost:50051"

	// Start the gRPC server in a goroutine
	go func() {
		log.Println("Starting gRPC server...")
		if err := server.Start(port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait a moment for the server to start
	time.Sleep(time.Second)

	// Create a gRPC client
	log.Println("Creating gRPC client...")
	c, err := client.NewClient(serverAddress)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	// Call the SayHello method
	name := "Gopher"
	log.Printf("Sending request to server with name: %s", name)
	message, err := c.SayHello(name)
	if err != nil {
		log.Fatalf("Error calling SayHello: %v", err)
	}

	log.Printf("Response from server: %s", message)

	// Keep the application running until interrupted
	log.Println("Server is running. Press Ctrl+C to stop.")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down...")
}
```

### Последовательность выполнения:

1. Запуск gRPC сервера в отдельной горутине
2. Ожидание запуска сервера (1 секунда)
3. Создание gRPC клиента
4. Отправка запроса "SayHello" с именем "Gopher"
5. Вывод полученного ответа
6. Ожидание сигнала прерывания (Ctrl+C) для завершения

### Вывод при запуске:

```
2025/03/13 14:59:58 Starting gRPC server...
2025/03/13 14:59:58 Server listening on port 50051
2025/03/13 14:59:59 Creating gRPC client...
2025/03/13 14:59:59 Sending request to server with name: Gopher
2025/03/13 14:59:59 Received request from: Gopher
2025/03/13 14:59:59 Response from server: Hello, Gopher!
2025/03/13 14:59:59 Server is running. Press Ctrl+C to stop.
^C2025/03/13 15:00:10 Shutting down...
```

## Обработка ошибок в gRPC

gRPC имеет встроенную систему обработки ошибок, основанную на статус-кодах. В нашем примере мы добавили метод `ValidateAndGreet`, который демонстрирует различные типы ошибок.

### Статус-коды gRPC

gRPC использует стандартный набор статус-кодов для обозначения различных типов ошибок:

| Код | Название            | Описание                             |
| --- | ------------------- | ------------------------------------ |
| 0   | OK                  | Успешное выполнение                  |
| 3   | INVALID_ARGUMENT    | Клиент указал недопустимый аргумент  |
| 5   | NOT_FOUND           | Запрошенный ресурс не найден         |
| 7   | PERMISSION_DENIED   | У клиента нет разрешения             |
| 8   | RESOURCE_EXHAUSTED  | Исчерпание ресурсов (квоты, лимиты)  |
| 9   | FAILED_PRECONDITION | Не выполнены предварительные условия |
| 10  | ABORTED             | Операция была прервана               |
| 13  | INTERNAL            | Внутренняя ошибка сервера            |
| 14  | UNAVAILABLE         | Сервис недоступен                    |
| 4   | DEADLINE_EXCEEDED   | Превышен срок ожидания               |

### Генерация ошибок на сервере

В нашем примере сервер генерирует различные типы ошибок в зависимости от входных данных:

```go
// Validate username
if username == "" {
    // Return INVALID_ARGUMENT error with details
    return nil, status.Errorf(codes.InvalidArgument, "username cannot be empty")
}

if strings.Contains(username, " ") {
    // Return INVALID_ARGUMENT error with details
    return nil, status.Errorf(codes.InvalidArgument, "username cannot contain spaces")
}

// Validate age
if age < 0 {
    return nil, status.Errorf(codes.InvalidArgument, "age cannot be negative")
}

if age < 18 {
    // Return PERMISSION_DENIED error
    return nil, status.Errorf(codes.PermissionDenied, "users under 18 are not allowed")
}

if age > 120 {
    // Return FAILED_PRECONDITION error
    return nil, status.Errorf(codes.FailedPrecondition, "age value %d seems unrealistic", age)
}
```

Для создания ошибок используется функция `status.Errorf()`, которая принимает код ошибки и сообщение.

### Обработка ошибок на клиенте

На стороне клиента мы можем извлечь статус-код и сообщение из ошибки:

```go
if err != nil {
    // Get the status from the error
    st, ok := status.FromError(err)
    if !ok {
        // Not a gRPC error
        return "", fmt.Errorf("unknown error: %v", err)
    }

    // Handle different error codes
    switch st.Code() {
    case codes.InvalidArgument:
        return "", fmt.Errorf("validation error: %s", st.Message())
    case codes.PermissionDenied:
        return "", fmt.Errorf("permission denied: %s", st.Message())
    case codes.FailedPrecondition:
        return "", fmt.Errorf("failed precondition: %s", st.Message())
    case codes.Internal:
        return "", fmt.Errorf("server internal error: %s", st.Message())
    case codes.DeadlineExceeded:
        return "", fmt.Errorf("request timed out: %s", st.Message())
    default:
        return "", fmt.Errorf("unexpected error (code=%s): %s", st.Code(), st.Message())
    }
}
```

Функция `status.FromError()` извлекает статус из ошибки gRPC, после чего мы можем проверить код ошибки и сообщение.

### Пример вывода с ошибками

При запуске приложения с различными входными данными мы получим следующие результаты:

```
--- Testing ValidateAndGreet with username=validuser, age=25 ---
2025/03/13 16:30:00 Validating request: username=validuser, age=25
2025/03/13 16:30:00 Validation passed for user: validuser
2025/03/13 16:30:00 Success: Welcome, validuser! You are 25 years old.

--- Testing ValidateAndGreet with username=, age=30 ---
2025/03/13 16:30:00 Validating request: username=, age=30
2025/03/13 16:30:00 Error: validation error: username cannot be empty

--- Testing ValidateAndGreet with username=user with spaces, age=35 ---
2025/03/13 16:30:00 Validating request: username=user with spaces, age=35
2025/03/13 16:30:00 Error: validation error: username cannot contain spaces

--- Testing ValidateAndGreet with username=younguser, age=15 ---
2025/03/13 16:30:00 Validating request: username=younguser, age=15
2025/03/13 16:30:00 Error: permission denied: users under 18 are not allowed

--- Testing ValidateAndGreet with username=olduser, age=150 ---
2025/03/13 16:30:00 Validating request: username=olduser, age=150
2025/03/13 16:30:00 Error: failed precondition: age value 150 seems unrealistic

--- Testing ValidateAndGreet with username=database_error, age=30 ---
2025/03/13 16:30:00 Validating request: username=database_error, age=30
2025/03/13 16:30:00 Error: server internal error: database connection failed

--- Testing ValidateAndGreet with username=timeout, age=30 ---
2025/03/13 16:30:00 Validating request: username=timeout, age=30
2025/03/13 16:30:00 Error: request timed out: operation timed out
```

### Расширенная обработка ошибок

В более сложных приложениях можно использовать дополнительные возможности:

1. **Детали ошибок** - добавление структурированных деталей к ошибкам:

```go
st := status.New(codes.InvalidArgument, "validation failed")
detailedStatus, err := st.WithDetails(
    &errdetails.BadRequest{
        FieldViolations: []*errdetails.BadRequest_FieldViolation{
            {
                Field: "username",
                Description: "cannot be empty",
            },
        },
    },
)
return nil, detailedStatus.Err()
```

2. **Перехватчики для централизованной обработки ошибок**:

```go
func errorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    resp, err := handler(ctx, req)
    if err != nil {
        // Log the error
        log.Printf("Error in %s: %v", info.FullMethod, err)
        // You could also modify the error here
    }
    return resp, err
}
```

3. **Повторные попытки при временных ошибках**:

```go
for retries := 0; retries < maxRetries; retries++ {
    resp, err := client.SomeMethod(ctx, req)
    if err != nil {
        st, ok := status.FromError(err)
        if !ok || (st.Code() != codes.Unavailable && st.Code() != codes.DeadlineExceeded) {
            return nil, err // Non-retryable error
        }
        // Wait before retry
        time.Sleep(backoff(retries))
        continue
    }
    return resp, nil
}
```

## Дополнительные возможности

В этом примере реализован только базовый функционал gRPC. В реальных приложениях можно использовать дополнительные возможности:

### 1. Безопасное соединение (TLS)

Для защиты соединения можно использовать TLS:

```go
// Сервер
creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
s := grpc.NewServer(grpc.Creds(creds))

// Клиент
creds, err := credentials.NewClientTLSFromFile("server.crt", "")
conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
```

### 2. Потоковая передача данных (Streaming)

gRPC поддерживает четыре типа методов:

- Унарный (Unary): клиент отправляет один запрос и получает один ответ
- Серверный стриминг: клиент отправляет один запрос и получает поток ответов
- Клиентский стриминг: клиент отправляет поток запросов и получает один ответ
- Двунаправленный стриминг: клиент и сервер обмениваются потоками сообщений

Пример определения в proto-файле:

```protobuf
service Greeter {
  // Унарный
  rpc SayHello (HelloRequest) returns (HelloReply) {}

  // Серверный стриминг
  rpc SayHellos (HelloRequest) returns (stream HelloReply) {}

  // Клиентский стриминг
  rpc SayManyHellos (stream HelloRequest) returns (HelloReply) {}

  // Двунаправленный стриминг
  rpc SayHelloToEveryone (stream HelloRequest) returns (stream HelloReply) {}
}
```

### 3. Middleware (Interceptors)

Для добавления общей функциональности (логирование, аутентификация, метрики) можно использовать перехватчики:

```go
// Серверный перехватчик
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    log.Printf("Request: %v", req)
    resp, err := handler(ctx, req)
    log.Printf("Response: %v", resp)
    return resp, err
}

s := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))

// Клиентский перехватчик
func loggingInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
    log.Printf("Before call: %s", method)
    err := invoker(ctx, method, req, reply, cc, opts...)
    log.Printf("After call: %s", method)
    return err
}

conn, err := grpc.Dial(address, grpc.WithUnaryInterceptor(loggingInterceptor))
```

### 4. Балансировка нагрузки

gRPC поддерживает балансировку нагрузки между несколькими серверами:

```go
conn, err := grpc.Dial(
    "service-name",
    grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
    grpc.WithResolvers(resolver),
)
```

### 5. Отказоустойчивость

Для повышения отказоустойчивости можно использовать повторные попытки и таймауты:

```go
// Таймаут
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})

// Повторные попытки
backoff := backoff.NewExponentialBackOff()
err := backoff.Retry(func() error {
    _, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
    return err
}, backoff)
```

## Заключение

gRPC предоставляет мощный и эффективный способ коммуникации между сервисами. Этот пример демонстрирует базовую реализацию клиент-серверного взаимодействия с использованием gRPC в Go.

Для более сложных сценариев рекомендуется изучить дополнительные возможности gRPC, такие как стриминг, перехватчики, балансировка нагрузки и отказоустойчивость.
