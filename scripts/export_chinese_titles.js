/**
 * Export tiêu đề tiếng Trung ra file text
 * Format: id|title mỗi dòng 1 entry
 */

const mysql = require('mysql2/promise');
const fs = require('fs');

const dbConfig = {
    host: 'localhost',
    user: 'root',
    password: '',
    database: 'mimirprompt_db'
};

async function main() {
    console.log('🔌 Connecting to database...');
    const connection = await mysql.createConnection(dbConfig);

    console.log('📥 Fetching Chinese titles...');
    const [rows] = await connection.execute(`
        SELECT id, title FROM prompts 
        WHERE title REGEXP '[一-龥]'
        ORDER BY id
    `);

    console.log(`📝 Found ${rows.length} Chinese titles`);

    // Create output with id|title per line
    const output = rows.map(row => `${row.id}|${row.title}`).join('\n');

    fs.writeFileSync('chinese_titles.txt', output, 'utf8');

    console.log('✅ Exported to chinese_titles.txt');
    console.log(`   Total: ${rows.length} titles`);

    await connection.end();
}

main().catch(console.error);
