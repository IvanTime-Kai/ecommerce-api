# Vocabulary

## S

**Segregation**
Tách biệt, phân tách trách nhiệm. Dùng trong CQRS: Command và Query được tách ra, không chồng chéo. Giống như phòng kế toán (ghi) và phòng báo cáo (đọc) là 2 phòng riêng.

## C

**Consistent / Consistency**
Nhất quán — data đọc ra phản ánh đúng state thực tế.

**Strong Consistency**
Nhất quán ngay lập tức. Sau khi ghi xong, đọc ra data đúng ngay. Dùng transaction trong DB. Ví dụ: số dư tài khoản ngân hàng.

**Eventual Consistency**
Nhất quán cuối cùng. Data sẽ đúng, nhưng không đúng ngay lập tức — có độ trễ nhỏ (vài ms đến vài giây). Chấp nhận được khi user không nhận ra sự khác biệt. Ví dụ: doanh thu báo cáo seller, số like trên Facebook.

## A

**Aggregate**
Tổng hợp nhiều row thành 1 kết quả. Trong SQL: `SUM`, `COUNT`, `AVG`, `GROUP BY` đều là aggregate. Vấn đề: bảng lớn → mỗi request aggregate lại → chậm. CQRS giải quyết bằng pre-aggregate — tính sẵn, lưu vào bảng riêng.
