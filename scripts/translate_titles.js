/**
 * Script dịch tiêu đề tiếng Trung sang tiếng Anh
 * Với retry logic và delay dài hơn để tránh rate limit
 */

const mysql = require('mysql2/promise');
const translate = require('@vitalets/google-translate-api');

const dbConfig = {
    host: 'localhost',
    user: 'root',
    password: '',
    database: 'mimirprompt_db'
};

// Delay helper
const delay = (ms) => new Promise(resolve => setTimeout(resolve, ms));

// Translate với retry
async function translateWithRetry(text, retries = 3) {
    for (let i = 0; i < retries; i++) {
        try {
            const result = await translate.translate(text, { from: 'zh-cn', to: 'en' });
            return result.text;
        } catch (error) {
            if (error.message.includes('Too Many Requests') && i < retries - 1) {
                console.log(`    ⏳ Rate limited, waiting 30s before retry ${i + 2}/${retries}...`);
                await delay(30000); // Wait 30s on rate limit
            } else {
                throw error;
            }
        }
    }
}

async function translateTitle(text) {
    try {
        // Check if text contains Chinese characters
        const chineseRegex = /[\u4e00-\u9fff]/g;
        const matches = text.match(chineseRegex);
        if (!matches || matches.length < 2) {
            return null; // Not Chinese, skip
        }

        return await translateWithRetry(text);
    } catch (error) {
        console.error(`    ❌ Error: ${error.message}`);
        return null;
    }
}

async function main() {
    console.log('🔌 Connecting to database...');
    const connection = await mysql.createConnection(dbConfig);

    console.log('📥 Fetching untranslated prompts...');
    // Only get prompts that still have Chinese characters
    const [rows] = await connection.execute(`
        SELECT id, title FROM prompts 
        WHERE title REGEXP '[一-龥]'
        ORDER BY id
    `);

    console.log(`📝 Found ${rows.length} prompts with Chinese titles left to translate\n`);

    let translated = 0;
    let skipped = 0;

    for (let i = 0; i < rows.length; i++) {
        const row = rows[i];
        console.log(`[${i + 1}/${rows.length}] ID ${row.id}: ${row.title.substring(0, 40)}...`);

        const newTitle = await translateTitle(row.title);

        if (newTitle && newTitle !== row.title) {
            console.log(`    ✅ → ${newTitle.substring(0, 50)}...`);
            await connection.execute('UPDATE prompts SET title = ? WHERE id = ?', [newTitle, row.id]);
            translated++;
        } else {
            console.log(`    ⚠ Skipped`);
            skipped++;
        }

        // Longer delay to avoid rate limit (3 seconds)
        await delay(3000);
    }

    console.log('\n✅ Translation complete!');
    console.log(`   Translated: ${translated}`);
    console.log(`   Skipped: ${skipped}`);

    await connection.end();
}

main().catch(console.error);
