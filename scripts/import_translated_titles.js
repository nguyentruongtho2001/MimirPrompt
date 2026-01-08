/**
 * Import translated English titles back to database
 * Maps line-by-line with original chinese_titles.txt to get IDs
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
    console.log('📖 Reading files...');

    // Read original file with IDs
    const originalContent = fs.readFileSync('chinese_titles.txt', 'utf8');
    const originalLines = originalContent.trim().split('\n');

    // Read translated file
    const translatedContent = fs.readFileSync('chinese_titles_EN_final.txt', 'utf8');
    const translatedLines = translatedContent.trim().split('\n');

    console.log(`📝 Original: ${originalLines.length} lines`);
    console.log(`📝 Translated: ${translatedLines.length} lines`);

    if (originalLines.length !== translatedLines.length) {
        console.error('❌ Line count mismatch! Aborting.');
        return;
    }

    // Build update map
    const updates = [];
    for (let i = 0; i < originalLines.length; i++) {
        const [id, originalTitle] = originalLines[i].split('|');
        const translatedTitle = translatedLines[i].trim();

        if (translatedTitle) {
            updates.push({ id: parseInt(id), title: translatedTitle });
        }
    }

    console.log(`\n🔌 Connecting to database...`);
    const connection = await mysql.createConnection(dbConfig);

    console.log(`📤 Updating ${updates.length} titles...`);

    let updated = 0;
    for (const { id, title } of updates) {
        await connection.execute('UPDATE prompts SET title = ? WHERE id = ?', [title, id]);
        updated++;
        if (updated % 100 === 0) {
            console.log(`   Progress: ${updated}/${updates.length}`);
        }
    }

    console.log(`\n✅ Imported ${updated} translated titles!`);

    await connection.end();
}

main().catch(console.error);
