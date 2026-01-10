# 🏠 Hướng Dẫn Self-Host với Cloudflare Tunnel

Triển khai MimirPrompt trên laptop Ubuntu tại nhà với domain `mimirprompt.com`.

---

## ✨ Ưu Điểm Của Cloudflare Tunnel

- ✅ **Miễn phí** SSL certificate
- ✅ **Bypass NAT/Firewall** - không cần port forward
- ✅ **IP động OK** - tự động cập nhật
- ✅ **Bảo mật** - server không expose port ra internet
- ✅ **DDoS protection** từ Cloudflare

---

## Bước 1: Cài Docker Trên Laptop Ubuntu

```bash
# Cập nhật hệ thống
sudo apt update && sudo apt upgrade -y

# Cài Docker
curl -fsSL https://get.docker.com | sh

# Thêm user vào docker group
sudo usermod -aG docker $USER
newgrp docker

# Kiểm tra
docker --version
```

---

## Bước 2: Mua Domain & Thêm Vào Cloudflare

### 2.1 Mua domain
Mua `mimirprompt.com` tại [Porkbun](https://porkbun.com) hoặc bất kỳ đâu.

### 2.2 Thêm domain vào Cloudflare (MIỄN PHÍ)
1. Đăng ký tại [dash.cloudflare.com](https://dash.cloudflare.com)
2. Click **"Add a Site"** → Nhập `mimirprompt.com`
3. Chọn plan **Free**
4. Cloudflare sẽ cho bạn 2 nameservers, ví dụ:
   - `anna.ns.cloudflare.com`
   - `bob.ns.cloudflare.com`
5. Vào nơi mua domain → **Đổi Nameservers** thành 2 cái của Cloudflare
6. Đợi 5-30 phút để cập nhật

---

## Bước 3: Tạo Cloudflare Tunnel

### 3.1 Vào Cloudflare Zero Trust
1. Vào [one.dash.cloudflare.com](https://one.dash.cloudflare.com)
2. Chọn **Networks** → **Tunnels**
3. Click **"Create a tunnel"**
4. Đặt tên: `mimir-tunnel`
5. Chọn **Cloudflared** connector
6. **QUAN TRỌNG**: Copy **TUNNEL TOKEN** (dạng `eyJhIjoi...`) và lưu lại

### 3.2 Cấu hình Public Hostname
Trong trang tunnel, thêm 2 hostnames:

| Public Hostname | Service |
|-----------------|---------|
| `mimirprompt.com` | `http://frontend:4321` |
| `api.mimirprompt.com` | `http://pocketbase:8090` |

---

## Bước 4: Clone Code Về Laptop

```bash
cd ~
git clone https://github.com/YOUR_USERNAME/MimirPrompt.git
cd MimirPrompt
```

Hoặc copy từ Windows qua USB/mạng nội bộ.

---

## Bước 5: Tạo File .env

```bash
cd ~/MimirPrompt
nano .env
```

Thêm nội dung (paste token từ Bước 3):
```env
CLOUDFLARE_TUNNEL_TOKEN=eyJhIjoixxxxxx...
```

Lưu file: `Ctrl+O` → Enter → `Ctrl+X`

---

## Bước 6: Chạy Ứng Dụng

```bash
cd ~/MimirPrompt

# Build và chạy với Cloudflare Tunnel
docker compose -f docker-compose.tunnel.yml up -d --build

# Xem logs
docker compose -f docker-compose.tunnel.yml logs -f
```

---

## Bước 7: Kiểm Tra

1. Truy cập **https://mimirprompt.com** → Website
2. Truy cập **https://api.mimirprompt.com/_/** → PocketBase Admin

🎉 **Done!** Website của bạn đã online!

---

## 🔧 Các Lệnh Hữu Ích

```bash
# Xem logs
docker compose -f docker-compose.tunnel.yml logs -f

# Restart
docker compose -f docker-compose.tunnel.yml restart

# Dừng
docker compose -f docker-compose.tunnel.yml down

# Xem status
docker compose -f docker-compose.tunnel.yml ps

# Update code mới
git pull
docker compose -f docker-compose.tunnel.yml up -d --build
```

---

## ⚠️ Lưu Ý Quan Trọng

### Laptop cần chạy 24/7
```bash
# Tắt tự động sleep
sudo systemctl mask sleep.target suspend.target hibernate.target hybrid-sleep.target

# Giữ laptop chạy khi gập màn hình
sudo nano /etc/systemd/logind.conf
# Thêm dòng: HandleLidSwitch=ignore
sudo systemctl restart systemd-logind
```

### Kiểm tra tunnel status
Vào [one.dash.cloudflare.com](https://one.dash.cloudflare.com) → Tunnels → Xem status "Healthy"

---

## 📊 So Sánh Chi Phí

| Mục | Chi phí |
|-----|---------|
| Cloudflare | MIỄN PHÍ |
| Domain (.com) | ~$10/năm |
| Điện laptop | ~50k-100k/tháng |
| **Tổng** | **~60k-110k/tháng** |

So với VPS (~120k/tháng), bạn tiết kiệm được khoảng 50%!
