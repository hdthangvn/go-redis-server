go-redis-server - ghi chú học tập (lecture 1 đến 5)

File này tóm tắt khái niệm cốt lõi của từng lecture, không đi sâu từng dòng code.
Dùng để ôn lại nhanh, hoặc chuẩn bị trả lời phỏng vấn.


LECTURE 1 - TCP server đơn giản nhất (lecture1_tcp_server/)

Đề bài: chứng minh có thể mở 1 cổng, nhận kết nối, đọc 1 dòng, trả lời.

Cơ chế:
- ln.Accept() = mở cửa đón khách (kết nối TCP mới).
- Mỗi khách vào thì tạo ngay 1 goroutine mới (go handleConnection(conn)) để nói chuyện riêng với khách đó.
- handleConnection chỉ Read() 1 lần, trả lời, rồi đóng kết nối.

Vấn đề để lại: không giới hạn số goroutine - 1 triệu client kết nối là 1 triệu goroutine, có thể làm sập server.


LECTURE 2 - Giới hạn số goroutine bằng pool (lecture2_thread_pool/)

Đề bài: đừng để số goroutine tạo ra là vô hạn.

Cơ chế - "vé vào cửa" (semaphore) bằng channel:
- GoroutinePool bọc 1 channel rỗng chan struct{} có sức chứa N.
- Get() = xin 1 vé (đẩy 1 giá trị vào channel). Nếu đã đủ N vé đang phát ra thì bị chặn, phải chờ.
- Return() = trả vé (lấy 1 giá trị ra khỏi channel).
- pool.Get() được gọi trong vòng lặp main(), trước khi tạo goroutine, nên nếu pool đầy, cả vòng lặp Accept() bị đứng khựng, client mới phải chờ ở tầng OS (chưa được accept), còn client đã accept thì luôn được xử lý ngay.

Kết quả: tại một thời điểm chỉ có tối đa N goroutine xử lý cùng lúc, an toàn tài nguyên hơn lecture 1.

Vấn đề để lại: vẫn là mô hình 1 kết nối = 1 goroutine (chỉ giới hạn số lượng, chưa đổi cách tiếp cận); và mỗi kết nối vẫn chỉ nói được 1 câu rồi cúp máy.


LECTURE 3 - Giao thức RESP thật + nhiều lệnh trên 1 kết nối (lecture3_resp_multi_command/)

Đề bài:
1. Một kết nối phải nói chuyện được nhiều câu liên tục, không chỉ 1 câu rồi cúp.
2. Server phải hiểu giao thức RESP thật (định dạng Redis dùng), không chỉ đọc text tuỳ tiện.
3. Server phải lưu trữ dữ liệu thật (SET/GET), an toàn khi nhiều goroutine truy cập cùng lúc.

Cơ chế:
- handleConnection giờ có vòng lặp for bao quanh: đọc 1 lệnh, xử lý, trả lời, rồi quay lại đọc lệnh tiếp theo, không đóng kết nối.
- Tách 2 package riêng:
  - protocol: chỉ biết "đọc/viết đúng định dạng bytes" (ReadCommand, EncodeSimpleString...). Không biết SET nghĩa là gì.
  - command: chỉ biết "PING/SET/GET nghĩa là gì và phải làm gì". Không biết gì về socket hay bytes thô.
- ReadCommand hiểu 2 dạng: inline (gõ tay đơn giản, ví dụ "SET foo bar") và multibulk (chuẩn RESP thật, dùng bởi redis-cli/redis-benchmark).
- Dữ liệu lưu trong 1 map[string]string dùng chung, bọc sync.Mutex, vì nhiều goroutine (nhiều kết nối) có thể cùng đọc/ghi map này, cần khoá để tránh race condition (2 goroutine ghi cùng lúc có thể làm Go panic).
- Pool size tăng từ 1 (demo) lên 1024, nghĩ tới việc chịu tải thật.

Vấn đề để lại: vẫn là mô hình 1 kết nối = 1 goroutine (dù giới hạn 1024). Nếu có hàng chục ngàn kết nối, việc tạo/quản lý từng đó goroutine vẫn tốn tài nguyên. Ngoài ra sync.Mutex là khoá cho cả map, không phải khoá riêng từng key, nên mọi SET/GET đều phải xếp hàng qua 1 khoá duy nhất (điểm nghẽn hiệu năng tiềm ẩn).


LECTURE 4 - I/O multiplexing: bỏ goroutine-per-connection (lecture4_io_multiplexing/)

Đề bài: có cách nào để 1 thread duy nhất theo dõi được hàng ngàn kết nối cùng lúc, mà không cần tạo hàng ngàn goroutine?

Cơ chế - nhờ kernel theo dõi giúp (epoll trên Linux, kqueue trên macOS):
- Ví dụ để hình dung: thay vì người phục vụ tự đi hỏi từng bàn "sẵn sàng gọi món chưa?" (tốn công), người phục vụ nhờ quản lý (kernel) "bàn nào sẵn sàng thì báo em" - người phục vụ chỉ ngồi 1 chỗ, gọi Wait().
- Monitor(event) = "đăng ký" 1 socket (bằng file descriptor - số hiệu của nó) để kernel theo dõi giúp.
- Wait() = hỏi kernel "trong các socket đã đăng ký, cái nào đang sẵn sàng đọc?" - nếu chưa có gì thì ngủ (không tốn CPU) cho tới khi có.
- Event loop = vòng lặp for chạy mãi mãi: Wait() -> nhận danh sách socket sẵn sàng -> xử lý từng cái -> quay lại Wait(). Tất cả trên 1 thread duy nhất, không tạo goroutine cho từng kết nối.
- Socket "lắng nghe" (serverFd, chờ khách mới) cũng được đăng ký theo dõi y hệt các kết nối đã có. Khi nó "sẵn sàng đọc" nghĩa là có khách mới, lúc đó mới Accept() rồi đăng ký tiếp kết nối mới đó vào danh sách theo dõi.
- IOMultiplexer là 1 interface chung; epoll_linux.go/kqueue_macos.go là 2 cách hiện thực khác nhau theo OS (chọn bằng build tag), main.go không cần biết đang chạy OS nào.

So sánh nhanh với lecture 1-3:
- Ai theo dõi kết nối có dữ liệu mới: lecture 1-3 là mỗi goroutine tự Read(), tự chờ riêng. Lecture 4 là kernel theo dõi tất cả, báo cho 1 thread.
- Số "người" chạy song song: lecture 1-3 là N goroutine (N = số kết nối, giới hạn bởi pool). Lecture 4 là 1 thread duy nhất (event loop).
- Khi nhiều client gửi lệnh cùng lúc: lecture 1-3 là nhiều goroutine được đánh thức song song. Lecture 4 là 1 lần Wait() trả về danh sách, xử lý tuần tự trong vòng lặp.

Vấn đề để lại (đã biết trước, chưa xử lý): nếu 1 lệnh RESP bị cắt làm 2 lần đọc (TCP chia gói giữa chừng), code hiện tại không ghép lại - mỗi lần đọc tạo buffer mới, không giữ phần dữ liệu dở dang của lần đọc trước. Ít gặp khi test tay qua telnet/nc, nhưng là điểm chưa hoàn thiện so với server thật (Redis thật giữ 1 buffer riêng cho từng connection để ghép nối liền mạch).


LECTURE 5 - Protocol không-blocking + key hết hạn (lecture5_resp_protocol/)

Lưu ý: lecture này thực ra gộp 2 chủ đề riêng biệt, không phải 1 khái niệm duy nhất.

Chủ đề A - vá lỗ hổng của lecture 4 (lệnh bị cắt làm 2 lần đọc):
- Đề bài: lecture 4 đọc bao nhiêu byte kernel đưa cho là xử lý ngay bấy nhiêu. Nếu 1 lệnh RESP bị TCP chia làm 2 gói (nửa đầu tới trước, nửa sau tới sau), lecture 4 sẽ đọc thiếu và coi là lỗi.
- Cơ chế: mỗi kết nối có 1 vùng đệm riêng (map pending, connFd -> bytes chưa xử lý hết). Mỗi lần đọc thêm dữ liệu, nối vào phần dư của lần trước, rồi thử tách lệnh. Nếu chưa đủ dữ liệu để thành 1 lệnh hoàn chỉnh, ParseCommand trả về lỗi đặc biệt ErrIncomplete (nghĩa là "chưa xong, đừng huỷ, chờ thêm") thay vì báo lỗi thật, và không tiêu tốn byte nào - phần dữ liệu dở dang được giữ nguyên trong pending, chờ lần đọc kế tiếp nối thêm vào.
- Đã kiểm tra thực tế: gửi 1 lệnh SET bị cắt làm 2 lần gửi cách nhau 0.3 giây, server vẫn ghép đúng và trả lời chính xác.

Chủ đề B - key có thể tự hết hạn (TTL/expiry), giống Redis thật:
- Đề bài: Redis thật cho phép đặt 1 key chỉ tồn tại trong X giây rồi tự biến mất (dùng làm cache, session, v.v.). Lecture 3/4 chưa có khái niệm này - key tồn tại mãi mãi cho tới khi bị SET đè.
- Cơ chế: mỗi giá trị lưu trong store giờ có thêm "thời điểm hết hạn" (expireAt) đi kèm. Thêm các lệnh mới: SET ... EX giây / PX mili-giây (đặt hạn dùng), TTL/PTTL (hỏi còn bao lâu hết hạn), EXPIRE (đặt hạn dùng cho key đã có sẵn), DEL, EXISTS.
- Có 2 cách 1 key hết hạn biến mất: "lazy" (khi GET/TTL tình cờ đọc trúng key đã hết hạn, xoá luôn lúc đó) và "active" (ActiveExpireCycle - định kỳ tự quét 1 mẫu nhỏ các key, xoá bớt key đã hết hạn, kể cả khi không ai đọc tới chúng) - giống hệt cách Redis thật vận hành, để key hết hạn không nằm mãi trong bộ nhớ chỉ vì không ai động tới.
- Vì ActiveExpireCycle chạy định kỳ nhưng KHÔNG dùng goroutine/ticker riêng: main.go gọi Wait() của epoll/kqueue kèm 1 khoảng timeout (100ms) - nếu không có client nào gửi gì trong 100ms, Wait() tự trả về rỗng, và main.go nhân lúc đó gọi ActiveExpireCycle() ngay trên cùng 1 thread duy nhất. Nhờ vậy sync.Mutex có thể bỏ hẳn khỏi store - chỉ có đúng 1 goroutine (event loop) từng đụng vào dữ liệu, không còn ai để tranh chấp cùng lúc nữa.

Vấn đề để lại: main.go mỗi lần có dữ liệu mới đều ghép "phần dư cũ + dữ liệu mới" thành 1 slice mới (có thể phải cấp phát lại bộ nhớ) - chấp nhận được với tải nhẹ, nhưng chưa phải cách tối ưu nhất cho server chịu tải cực cao.


MẠCH TIẾN HOÁ XUYÊN SUỐT 5 LECTURE (điều quan trọng nhất để nhớ)

L1: 1 kết nối = 1 goroutine, KHÔNG giới hạn số lượng
  -> vấn đề: quá tải khi có quá nhiều kết nối
L2: 1 kết nối = 1 goroutine, CÓ giới hạn bằng pool (semaphore)
  -> vấn đề: mỗi kết nối chỉ nói được 1 câu, chưa hiểu giao thức thật
L3: giữ pool, nhưng đọc NHIỀU lệnh/kết nối + hiểu RESP protocol + lưu trữ có khoá
  -> vấn đề: vẫn tạo 1 goroutine/kết nối, không scale tốt khi có rất nhiều kết nối
L4: bỏ goroutine-per-connection, dùng 1 thread + epoll/kqueue theo dõi TẤT CẢ kết nối
  -> vấn đề: lệnh RESP bị cắt làm 2 lần đọc thì không ghép lại được; chưa có khái niệm key tự hết hạn
L5: vá lỗi ghép lệnh bị cắt (pending buffer + ErrIncomplete) + thêm TTL/expiry giống Redis thật

Mỗi lecture đều sinh ra từ 1 câu hỏi "cách hiện tại còn thiếu gì?" của lecture trước - đây là cách trả lời phỏng vấn tự nhiên nhất khi được hỏi "tại sao code lại thiết kế như vậy".
