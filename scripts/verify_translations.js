/**
 * Verify imported translations
 */

const mysql = require('mysql2/promise');

const dbConfig = {
    host: 'localhost',
    user: 'root',
    password: '',
    database: 'mimirprompt_db'
};

async function main() {
    const connection = await mysql.createConnection(dbConfig);

    // Count prompts with Chinese vs English titles
    const [chineseCount] = await connection.execute(`
        SELECT COUNT(*) as count FROM prompts WHERE title REGEXP '[一-龥]'
    `);

    const [totalCount] = await connection.execute(`
        SELECT COUNT(*) as count FROM prompts
    `);

    console.log('📊 Statistics:');
    console.log(`   Total prompts: ${totalCount[0].count}`);
    console.log(`   Still have Chinese: ${chineseCount[0].count}`);
    console.log(`   Translated to English: ${totalCount[0].count - chineseCount[0].count}`);

    // Show sample of translated titles
    console.log('\n📝 Sample of translated prompts:');
    const [samples] = await connection.execute(`
        SELECT id, title FROM prompts 
        WHERE title NOT REGEXP '[一-龥]'
        ORDER BY id
        LIMIT 10
    `);

    samples.forEach(s => {
        console.log(`   [${s.id}] ${s.title}`);
    });

    // Check if any still have Chinese
    if (chineseCount[0].count > 0) {
        console.log('\n⚠️ Prompts still with Chinese titles:');
        const [remaining] = await connection.execute(`
            SELECT id, title FROM prompts 
            WHERE title REGEXP '[一-龥]'
            LIMIT 5
        `);
        remaining.forEach(s => {
            console.log(`   [${s.id}] ${s.title}`);
        });
    }

    await connection.end();
}

main().catch(console.error);
