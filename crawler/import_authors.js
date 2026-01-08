const mysql = require('mysql2/promise');
const fs = require('fs');

// Extract author info from source URL
function extractAuthorFromUrl(sourceUrl) {
    if (!sourceUrl) return null;

    // Twitter/X URL pattern: https://x.com/username/status/...
    // or https://twitter.com/username/status/...
    const twitterPattern = /(?:twitter\.com|x\.com)\/([^\/]+)\/status/;
    const match = sourceUrl.match(twitterPattern);

    if (match) {
        const username = match[1];
        return {
            username: username,
            platform: 'twitter',
            profile_url: `https://x.com/${username}`,
            name: username // Will use username as name initially
        };
    }

    return null;
}

async function importAuthors() {
    console.log('🔧 Importing authors from prompts...\n');

    const connection = await mysql.createConnection({
        host: '127.0.0.1',
        port: 3306,
        user: 'root',
        password: '',
        database: 'mimirprompt_db',
        charset: 'utf8mb4'
    });

    try {
        // Get all prompts with source_url
        const [prompts] = await connection.execute(
            'SELECT id, source_url FROM prompts WHERE source_url IS NOT NULL AND source_url != ""'
        );

        console.log(`📊 Found ${prompts.length} prompts with source URLs\n`);

        // Extract unique authors
        const authorsMap = new Map();

        for (const prompt of prompts) {
            const authorInfo = extractAuthorFromUrl(prompt.source_url);
            if (authorInfo && !authorsMap.has(authorInfo.username)) {
                authorsMap.set(authorInfo.username, {
                    ...authorInfo,
                    promptIds: [prompt.id]
                });
            } else if (authorInfo) {
                authorsMap.get(authorInfo.username).promptIds.push(prompt.id);
            }
        }

        console.log(`👤 Found ${authorsMap.size} unique authors\n`);

        let imported = 0;
        let updated = 0;

        for (const [username, author] of authorsMap) {
            // Check if author already exists
            const [existing] = await connection.execute(
                'SELECT id FROM authors WHERE username = ?',
                [username]
            );

            let authorId;

            if (existing.length > 0) {
                authorId = existing[0].id;
                updated++;
            } else {
                // Insert new author
                const [result] = await connection.execute(
                    'INSERT INTO authors (name, username, platform, profile_url, prompt_count) VALUES (?, ?, ?, ?, ?)',
                    [author.name, author.username, author.platform, author.profile_url, author.promptIds.length]
                );
                authorId = result.insertId;
                imported++;
                console.log(`  ✅ Added: @${username} (${author.promptIds.length} prompts)`);
            }

            // Update prompts with author_id
            for (const promptId of author.promptIds) {
                await connection.execute(
                    'UPDATE prompts SET author_id = ? WHERE id = ?',
                    [authorId, promptId]
                );
            }
        }

        // Update prompt_count for all authors
        await connection.execute(`
            UPDATE authors a 
            SET prompt_count = (SELECT COUNT(*) FROM prompts p WHERE p.author_id = a.id)
        `);

        console.log(`\n✅ Import complete!`);
        console.log(`   New authors: ${imported}`);
        console.log(`   Updated: ${updated}`);

    } catch (error) {
        console.error('❌ Error:', error.message);
    } finally {
        await connection.end();
    }
}

importAuthors();
