package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

// Số lượng worker cố định trong pool.
// Đây là điểm khác biệt cốt lõi so với "1 connection = 1 goroutine":
// dù có 10.000 client kết nối, ta chỉ có đúng WORKER_COUNT goroutine xử lý cùng lúc.
const workerCount = 3

// jobQueue là hàng đợi (channel) chứa các connection đang chờ được xử lý.
// Buffer size = 100 nghĩa là: có thể "xếp hàng" tối đa 100 connection
// trước khi Accept() bị chặn lại (backpressure).
var jobQueue = make(chan net.Conn, 100)

func main() {
	listener, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("Server (thread pool) đang lắng nghe tại :4000 ...")

	// Bước 1: Khởi tạo pool gồm workerCount goroutine.
	// Mỗi worker chạy vòng lặp riêng, tự lấy job từ jobQueue.
	for i := 1; i <= workerCount; i++ {
		go worker(i)
	}

	// Bước 2: Accept loop - vòng này KHÔNG xử lý connection trực tiếp,
	// nó chỉ nhận connection rồi "ném" vào hàng đợi cho worker xử lý.
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}

		// Nếu jobQueue đầy (100 connection đang chờ), dòng này sẽ BLOCK
		// cho đến khi có worker rảnh lấy bớt job ra -> đây chính là cơ chế
		// giới hạn tải (backpressure), tránh server bị quá tải vô hạn.
		jobQueue <- conn
	}
}

// worker là một "luồng" cố định trong pool, sống suốt vòng đời server.
// Nó liên tục lấy connection từ jobQueue và xử lý tuần tự (từng cái một).
func worker(id int) {
	for conn := range jobQueue {
		fmt.Println("Worker", id, "đang xử lý:", conn.RemoteAddr())
		handleConnection(conn)
		fmt.Println("Worker", id, "đã xử lý xong:", conn.RemoteAddr())
	}
}

// handleConnection giống hệt bản echo server gốc: đọc từng dòng và echo lại.
func handleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		request := scanner.Text()
		fmt.Println("Nhận từ", remoteAddr, ":", request)

		response := "echo: " + request + "\n"
		if _, err := io.WriteString(conn, response); err != nil {
			log.Println("write error:", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("read error:", err)
	}
	fmt.Println("Client ngắt kết nối:", remoteAddr)
}
