# 🚀 Hướng Dẫn Triển Khai MimirPrompt

Hướng dẫn từng bước triển khai lên VPS với domain `mimirprompt.com`.

---

## 📋 Yêu Cầu

- VPS với Ubuntu 22.04+ (tối thiểu 2GB RAM)
- Domain đã trỏ về IP của VPS
- Tài khoản SSH

---

## Bước 1: Thuê VPS

**Gợi ý nhà cung cấp rẻ:**
- [Vultr](https://vultr.com) - $5/tháng (1GB RAM) hoặc $10/tháng (2GB RAM)
- [DigitalOcean](https://digitalocean.com) - $6/tháng
- [Hetzner](https://hetzner.com) - €4/tháng (rẻ nhất)

**Cấu hình khuyến nghị:**
- OS: Ubuntu 22.04 LTS
- RAM: 2GB trở lên
- Storage: 50GB SSD

---

## Bước 2: Mua Domain & Cấu Hình DNS

1. Mua domain `mimirprompt.com` tại [Namecheap](https://namecheap.com), [Porkbun](https://porkbun.com), hoặc [Cloudflare](https://cloudflare.com)

2. Thêm DNS Records (thay `YOUR_VPS_IP` bằng IP thật):

| Type | Name | Value |
|------|------|-------|
| A | @ | YOUR_VPS_IP |
| A | www | YOUR_VPS_IP |
| A | api | YOUR_VPS_IP |

3. Đợi DNS propagate (5-30 phút)

---

## Bước 3: Cài Đặt Server

SSH vào VPS và chạy:

```bash
# Cập nhật hệ thống
sudo apt update && sudo apt upgrade -y

# Cài Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Cài Docker Compose
sudo apt install docker-compose-plugin -y

# Thêm user vào docker group (không cần sudo)
sudo usermod -aG docker $USER
newgrp docker
```

---

## Bước 4: Clone Project

```bash
# Clone repo
cd ~
git clone https://github.com/YOUR_USERNAME/MimirPrompt.git
cd MimirPrompt

# (Nếu chưa push lên GitHub, dùng SCP để upload files)
```

**Nếu chưa có GitHub repo, upload bằng SCP từ máy local:**

```powershell
# Chạy trên Windows (PowerShell)
scp -r C:\Users\Thor\Desktop\MimirPrompt root@YOUR_VPS_IP:~/
```

---

## Bước 5: Lấy SSL Certificate (Lần Đầu)

```bash
cd ~/MimirPrompt

# Dùng nginx config tạm
cp nginx/nginx.initial.conf nginx/nginx.conf.bak
mv nginx/nginx.initial.conf nginx/nginx.conf

# Chạy nginx tạm
docker compose up -d nginx

# Lấy SSL certificate
docker compose run --rm certbot certonly \
  --webroot \
  --webroot-path=/var/www/certbot \
  -d mimirprompt.com \
  -d www.mimirprompt.com \
  -d api.mimirprompt.com \
  --email your-email@example.com \
  --agree-tos \
  --no-eff-email

# Khôi phục nginx config chính
mv nginx/nginx.conf.bak nginx/nginx.conf

# Dừng nginx tạm
docker compose down
```

---

## Bước 6: Chạy Ứng Dụng

```bash
cd ~/MimirPrompt

# Build và chạy tất cả services
docker compose up -d --build

# Kiểm tra logs
docker compose logs -f

# Kiểm tra status
docker compose ps
```

---

## Bước 7: Kiểm Tra

1. Truy cập https://mimirprompt.com - Website chính
2. Truy cập https://api.mimirprompt.com/_/ - PocketBase Admin UI
3. Tạo tài khoản admin cho PocketBase

---

## 🔧 Các Lệnh Hữu Ích

```bash
# Xem logs
docker compose logs -f frontend
docker compose logs -f pocketbase

# Restart services
docker compose restart

# Dừng tất cả
docker compose down

# Cập nhật code mới
git pull
docker compose up -d --build

# Backup database
docker compose exec pocketbase ./pocketbase backup
```

---

## ⚠️ Troubleshooting

**Lỗi SSL certificate:**
```bash
# Xóa và lấy lại certificate
sudo rm -rf nginx/ssl/live/mimirprompt.com
# Chạy lại Bước 5
```

**Lỗi port đã được sử dụng:**
```bash
sudo lsof -i :80
sudo lsof -i :443
# Kill process đang dùng port
```

**Kiểm tra firewall:**
```bash
sudo ufw allow 80
sudo ufw allow 443
sudo ufw allow 22
sudo ufw enable
```

---

## 📊 Chi Phí Ước Tính

| Mục | Chi phí |
|-----|---------|
| VPS (Hetzner CX21) | ~€4/tháng |
| Domain (.com) | ~$10-12/năm |
| SSL | FREE (Let's Encrypt) |
| **Tổng** | **~$6-7/tháng** |
