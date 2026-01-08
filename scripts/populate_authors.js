/**
 * Extract authors from prompts.json and update database
 * Gets Twitter username from sourceUrl and maps to authors table
 */

const mysql = require('mysql2/promise');
const fs = require('fs');

const dbConfig = {
    host: 'localhost',
    user: 'root',
    password: '',
    database: 'mimirprompt_db'
};

function extractTwitterUsername(url) {
    if (!url) return null;
    // Match patterns like https://x.com/username/... or https://twitter.com/username/...
    const match = url.match(/(?:x\.com|twitter\.com)\/([^\/\?]+)/i);
    return match ? match[1] : null;
}

async function main() {
    console.log('📖 Reading prompts.json...');
    const data = JSON.parse(fs.readFileSync('../data/prompts.json', 'utf8'));
    const prompts = data.prompts;

    console.log(`📝 Found ${prompts.length} prompts`);

    // Extract unique usernames
    const usernameMap = new Map();
    for (const p of prompts) {
        const username = extractTwitterUsername(p.sourceUrl);
        if (username) {
            usernameMap.set(username.toLowerCase(), username);
        }
    }

    console.log(`👤 Found ${usernameMap.size} unique Twitter usernames`);

    // Connect to database
    console.log('\n🔌 Connecting to database...');
    const connection = await mysql.createConnection(dbConfig);

    // Get existing authors
    const [existingAuthors] = await connection.execute('SELECT id, username, name FROM authors');
    const authorMap = new Map();
    for (const a of existingAuthors) {
        if (a.username) authorMap.set(a.username.toLowerCase(), a.id);
        if (a.name) authorMap.set(a.name.toLowerCase(), a.id);
    }

    console.log(`📊 Existing authors: ${existingAuthors.length}`);

    // Insert new authors
    let newAuthors = 0;
    for (const [lower, original] of usernameMap) {
        if (!authorMap.has(lower)) {
            const [result] = await connection.execute(
                'INSERT INTO authors (username, name, profile_url) VALUES (?, ?, ?)',
                [original, original, `https://x.com/${original}`]
            );
            authorMap.set(lower, result.insertId);
            newAuthors++;
        }
    }
    console.log(`➕ Added ${newAuthors} new authors`);

    // Update prompts with author_id based on title matching
    console.log('\n📤 Updating prompts with author_id...');

    // Get all prompts from database with their titles
    const [dbPrompts] = await connection.execute('SELECT id, title FROM prompts');

    let updated = 0;
    for (const dbPrompt of dbPrompts) {
        // Find matching prompt in JSON by matching case number
        const caseMatch = dbPrompt.title.match(/Case (\d+)/i);
        if (!caseMatch) continue;
        const caseNum = parseInt(caseMatch[1]);

        // Find corresponding JSON prompt (857 - index = case_number)
        const jsonPrompt = prompts.find(p => {
            const jsonCase = p.title.match(/案例 (\d+)/);
            return jsonCase && parseInt(jsonCase[1]) === caseNum;
        });

        if (!jsonPrompt) continue;

        const username = extractTwitterUsername(jsonPrompt.sourceUrl);
        if (!username) continue;

        const authorId = authorMap.get(username.toLowerCase());
        if (!authorId) continue;

        await connection.execute(
            'UPDATE prompts SET author_id = ? WHERE id = ?',
            [authorId, dbPrompt.id]
        );
        updated++;
    }

    console.log(`✅ Updated ${updated} prompts with author_id`);

    // Show sample
    const [samples] = await connection.execute(`
        SELECT p.id, p.title, a.username 
        FROM prompts p 
        LEFT JOIN authors a ON p.author_id = a.id 
        WHERE p.author_id IS NOT NULL 
        LIMIT 5
    `);
    console.log('\n📝 Sample results:');
    samples.forEach(s => console.log(`   [${s.id}] ${s.title.substring(0, 40)}... → @${s.username}`));

    await connection.end();
}

main().catch(console.error);
