# MimirPrompt - AI Prompt Gallery

Thư viện AI Prompts được crawl từ opennana.com

## Cấu trúc thư mục

```
MimirPrompt/
├── crawler/           # Playwright crawler script
│   ├── package.json
│   ├── crawler.js
│   └── node_modules/
├── data/              # Dữ liệu đã crawl
│   └── prompts.json   # 857 prompts (~2MB)
├── backend/           # Go backend API
│   ├── config/
│   │   └── migration_prompts.sql
│   ├── api/
│   ├── models/
│   └── importer/
└── frontend/          # Admin UI (Astro/React)
```

## Dữ liệu đã crawl

- **Tổng số prompts:** 857
- **Prompts có text:** 852
- **Nguồn:** https://opennana.com/awesome-prompt-gallery/

### Cấu trúc mỗi prompt:
```json
{
  "index": 0,
  "title": "案例 857：超逼真的3D商业风格产品图",
  "thumbnail": "https://opennana.com/.../857.jpeg",
  "sourceUrl": "https://x.com/...",
  "images": ["..."],
  "tags": ["landscape", "nature", "photography", "product"],
  "promptText": "Create an ultra-realistic 3D...",
  "promptCount": 2
}
```

## Hướng dẫn sử dụng

### 1. Chạy lại crawler (nếu cần)
```bash
cd crawler
npm install
npx playwright install chromium
node crawler.js
```

### 2. Setup database
```bash
mysql -u root -p < backend/config/migration_prompts.sql
```

### 3. Import data
```bash
cd backend
go run cmd/import/main.go
```

### 4. Chạy API server
```bash
cd backend
go run cmd/api/main.go
```

## Tags có sẵn

- 3d, animal, architecture, branding, cartoon
- character, clay, creative, data-viz, emoji
- fantasy, fashion, felt, food, futuristic
- gaming, illustration, infographic, interior
- landscape, logo, minimalist, nature, neon
- paper-craft, photography, pixel, portrait
- poster, product, retro, sci-fi, sculpture
- toy, typography, ui, vehicle
