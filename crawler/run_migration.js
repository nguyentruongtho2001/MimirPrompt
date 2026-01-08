const mysql = require('mysql2/promise');

async function runMigration() {
    console.log('🔧 Running authors migration...\n');

    const connection = await mysql.createConnection({
        host: '127.0.0.1',
        port: 3306,
        user: 'root',
        password: '',
        database: 'mimirprompt_db',
        charset: 'utf8mb4',
        multipleStatements: true
    });

    try {
        // Create authors table
        await connection.execute(`
            CREATE TABLE IF NOT EXISTS authors (
                id INT AUTO_INCREMENT PRIMARY KEY,
                name VARCHAR(255) NOT NULL,
                username VARCHAR(100),
                platform VARCHAR(50) DEFAULT 'twitter',
                profile_url VARCHAR(500),
                avatar_url VARCHAR(500),
                bio TEXT,
                prompt_count INT DEFAULT 0,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                INDEX idx_username (username),
                INDEX idx_platform (platform),
                INDEX idx_prompt_count (prompt_count)
            )
        `);
        console.log('✅ Created authors table');

        // Check if author_id column exists
        const [columns] = await connection.execute(`
            SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS 
            WHERE TABLE_SCHEMA = 'mimirprompt_db' 
            AND TABLE_NAME = 'prompts' 
            AND COLUMN_NAME = 'author_id'
        `);

        if (columns.length === 0) {
            await connection.execute(`
                ALTER TABLE prompts 
                ADD COLUMN author_id INT DEFAULT NULL AFTER source_url,
                ADD INDEX idx_author_id (author_id)
            `);
            console.log('✅ Added author_id column to prompts');
        } else {
            console.log('ℹ️  author_id column already exists');
        }

        console.log('\n🎉 Migration complete!');

    } catch (error) {
        console.error('❌ Error:', error.message);
    } finally {
        await connection.end();
    }
}

runMigration();
