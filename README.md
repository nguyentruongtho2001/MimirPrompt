# MimirPrompt - AI Prompt Gallery

Thư viện AI Prompts được crawl từ opennana.com, xây dựng với Astro + PocketBase.

## 🚀 Demo

- **Frontend:** http://localhost:4321
- **PocketBase Admin:** http://127.0.0.1:8090/_/

## 📁 Cấu trúc thư mục

```
MimirPrompt/
├── crawler/              # Playwright crawler script
│   ├── crawler.js
│   └── package.json
├── data/                 # Dữ liệu đã crawl
│   └── prompts.json      # 857 prompts (~2MB)
├── pocketbase/           # PocketBase backend
│   ├── pocketbase.exe
│   ├── pb_data/
│   └── pb_migrations/
└── frontend/             # Astro frontend
    ├── src/
    │   ├── components/   # UI Components
    │   ├── pages/        # Routes
    │   ├── layouts/      # Layout templates
    │   └── lib/          # PocketBase client
    └── public/
```

## ✨ Tính năng

- 🔍 **Tìm kiếm** prompts theo tiêu đề
- 🏷️ **Tags & Categories** để lọc prompts
- 🎠 **Spotlight Carousel** cho prompts mới nhất
- 📌 **Bookmark** lưu prompts yêu thích (localStorage)
- 🎲 **Random Prompt** khám phá ngẫu nhiên
- 📋 **Copy Prompt** với 1 click
- 🌙 Neo-Brutalism design style

## 🛠️ Hướng dẫn cài đặt

### 1. Chạy PocketBase
```bash
cd pocketbase
.\pocketbase.exe serve
```

### 2. Import dữ liệu (lần đầu)
```bash
cd pocketbase
node migrate.js
```

### 3. Chạy Frontend
```bash
cd frontend
npm install
npm run dev
```

## 📊 Dữ liệu

- **Tổng số prompts:** 857
- **Prompts có text:** 852
- **Nguồn:** https://opennana.com/awesome-prompt-gallery/

### Cấu trúc mỗi prompt:
```json
{
  "title": "案例 857：超逼真的3D商业风格产品图",
  "thumbnail": "https://opennana.com/.../857.jpeg",
  "source_url": "https://x.com/...",
  "images_list": ["..."],
  "prompt_text": "Create an ultra-realistic 3D...",
  "tags": ["product", "3d", "photography"],
  "category": "product"
}
```

## 🏷️ Tags có sẵn

`3d` `animal` `architecture` `branding` `cartoon` `character` `clay` `creative` `data-viz` `emoji` `fantasy` `fashion` `felt` `food` `futuristic` `gaming` `illustration` `infographic` `interior` `landscape` `logo` `minimalist` `nature` `neon` `paper-craft` `photography` `pixel` `portrait` `poster` `product` `retro` `sci-fi` `sculpture` `toy` `typography` `ui` `vehicle`

## 🔧 Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | Astro, TypeScript |
| Backend | PocketBase |
| Database | SQLite (embedded) |
| Styling | CSS (Neo-Brutalism) |

## 📝 License

MIT
