const mysql = require('mysql2/promise');

// Correct Vietnamese descriptions for tags
const tagDescriptions = {
    '3D': 'Thiết kế và render 3D',
    'Animal': 'Động vật',
    'Architecture': 'Kiến trúc',
    'Branding': 'Thương hiệu',
    'Cartoon': 'Hoạt hình',
    'Character': 'Nhân vật',
    'Clay': 'Phong cách đất sét',
    'Creative': 'Sáng tạo',
    'Data-Viz': 'Trực quan hóa dữ liệu',
    'Emoji': 'Biểu tượng cảm xúc',
    'Fantasy': 'Giả tưởng',
    'Fashion': 'Thời trang',
    'Felt': 'Phong cách len dạ',
    'Food': 'Đồ ăn',
    'Futuristic': 'Tương lai',
    'Gaming': 'Game',
    'Illustration': 'Minh họa',
    'Infographic': 'Đồ họa thông tin',
    'Interior': 'Nội thất',
    'Landscape': 'Phong cảnh',
    'Logo': 'Thiết kế logo',
    'Minimalist': 'Tối giản',
    'Nature': 'Thiên nhiên',
    'Neon': 'Phong cách neon',
    'Paper-Craft': 'Nghệ thuật giấy',
    'Photography': 'Nhiếp ảnh',
    'Pixel': 'Pixel art',
    'Portrait': 'Chân dung',
    'Poster': 'Áp phích',
    'Product': 'Sản phẩm',
    'Retro': 'Phong cách cổ điển',
    'Sci-Fi': 'Khoa học viễn tưởng',
    'Sculpture': 'Điêu khắc',
    'Toy': 'Đồ chơi',
    'Typography': 'Typography',
    'UI': 'Giao diện người dùng',
    'Vehicle': 'Phương tiện'
};

async function fixTagDescriptions() {
    console.log('🔧 Fixing tag descriptions encoding...');

    const connection = await mysql.createConnection({
        host: '127.0.0.1',
        port: 3306,
        user: 'root',
        password: '',
        database: 'mimirprompt_db',
        charset: 'utf8mb4'
    });

    try {
        for (const [name, description] of Object.entries(tagDescriptions)) {
            await connection.execute(
                'UPDATE prompt_tags SET description = ? WHERE name = ?',
                [description, name]
            );
            console.log(`  ✅ Updated: ${name} -> ${description}`);
        }

        console.log('\n✅ All tag descriptions have been fixed!');
    } catch (error) {
        console.error('❌ Error:', error.message);
    } finally {
        await connection.end();
    }
}

fixTagDescriptions();
