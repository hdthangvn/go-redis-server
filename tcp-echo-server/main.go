package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	// Bước 1: "Mở cổng" - tạo Listener lắng nghe trên port 4000.
	listener, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("Server đang lắng nghe tại :4000 ...")

	// Bước 2: Vòng lặp Accept - server sống mãi, liên tục chờ client mới.
	for {
		conn, err := listener.Accept() // block ở đây cho đến khi có client kết nối
		if err != nil {
			log.Println("accept error:", err)
			continue
		}

		// Bước 3: Mỗi connection xử lý trong 1 goroutine riêng,
		// để không chặn việc Accept() các client khác.
		go handleConnection(conn)
	}
}

// handleConnection đọc NHIỀU request từ CÙNG MỘT connection,
// cho đến khi client đóng kết nối (EOF) hoặc có lỗi.
func handleConnection(conn net.Conn) {
	defer conn.Close() // đảm bảo luôn đóng connection khi hàm kết thúc

	remoteAddr := conn.RemoteAddr()
	fmt.Println("Client kết nối:", remoteAddr)

	// bufio.Scanner giúp đọc dữ liệu theo từng dòng (tách bởi \n).
	// Đây là cách ta tự định nghĩa "ranh giới 1 request" trên luồng byte thô của TCP.
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() { // vòng lặp: mỗi lần lặp = 1 request (1 dòng) mới từ client
		request := scanner.Text()
		fmt.Printf("[%s] nhận: %q\n", remoteAddr, request)

		response := "echo: " + request + "\n"
		_, err := io.WriteString(conn, response)
		if err != nil {
			log.Println("write error:", err)
			return
		}
	}

	// scanner.Scan() trả về false khi: client đóng connection (EOF) hoặc có lỗi đọc.
	if err := scanner.Err(); err != nil {
		log.Println("read error:", err)
	}
	fmt.Println("Client ngắt kết nối:", remoteAddr)
}
