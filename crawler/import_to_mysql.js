/**
 * Import Prompts to MySQL Database
 * Uses local image paths instead of remote URLs
 */

const mysql = require('mysql2/promise');
const fs = require('fs');
const path = require('path');

// Configuration
const DB_CONFIG = {
    host: '127.0.0.1',
    port: 3306,
    user: 'root',
    password: '',
    database: 'mimirprompt_db',
    charset: 'utf8mb4'
};

const DATA_FILE = path.join(__dirname, '..', 'data', 'prompts.json');
const IMAGES_DIR = path.join(__dirname, '..', 'data', 'images');

// Image URL prefix for serving images (relative path for web)
const IMAGES_URL_PREFIX = '/images';

/**
 * Sanitize filename - same logic as download script
 */
function sanitizeFilename(name, maxLength = 50) {
    if (!name) return 'untitled';

    const caseMatch = name.match(/案例\s*(\d+)/);
    const caseNum = caseMatch ? caseMatch[1] : '';

    let cleaned = name
        .replace(/案例\s*\d+[：:]\s*/g, '')
        .replace(/[<>:"/\\|?*\x00-\x1F]/g, '')
        .replace(/\s+/g, '_')
        .replace(/_{2,}/g, '_')
        .trim();

    if (cleaned.length > maxLength) {
        cleaned = cleaned.substring(0, maxLength);
    }

    return caseNum ? `${caseNum}_${cleaned}` : cleaned;
}

/**
 * Get extension from URL
 */
function getExtension(url) {
    const match = url.match(/\.(\w+)(?:\?|$)/);
    return match ? match[1].toLowerCase() : 'jpg';
}

/**
 * Convert remote URL to local path
 */
function getLocalImagePath(url, title, imageIndex = 0) {
    const baseName = sanitizeFilename(title);
    const ext = getExtension(url);

    // Check which file exists
    const possibleNames = [
        `${baseName}.${ext}`,
        `${baseName}_${imageIndex + 1}.${ext}`,
        `${baseName}.jpeg`,
        `${baseName}.jpg`,
        `${baseName}.png`,
        `${baseName}_${imageIndex + 1}.jpeg`,
        `${baseName}_${imageIndex + 1}.jpg`,
        `${baseName}_${imageIndex + 1}.png`
    ];

    for (const name of possibleNames) {
        const fullPath = path.join(IMAGES_DIR, name);
        if (fs.existsSync(fullPath)) {
            return `${IMAGES_URL_PREFIX}/${name}`;
        }
    }

    // Fallback to expected name even if file doesn't exist
    const filename = imageIndex === 0 ? `${baseName}.${ext}` : `${baseName}_${imageIndex + 1}.${ext}`;
    return `${IMAGES_URL_PREFIX}/${filename}`;
}

/**
 * Extract case number from title
 */
function extractCaseNumber(title) {
    const match = title.match(/案例\s*(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
}

/**
 * Main import function
 */
async function importData() {
    console.log('🚀 MimirPrompt Database Importer');
    console.log('=================================\n');

    // Read prompts data
    console.log('📂 Reading prompts.json...');
    let data;
    try {
        const raw = fs.readFileSync(DATA_FILE, 'utf8');
        data = JSON.parse(raw);
    } catch (error) {
        console.error(`❌ Error reading data file: ${error.message}`);
        process.exit(1);
    }
    console.log(`✅ Found ${data.prompts.length} prompts\n`);

    // Connect to database
    console.log('🔌 Connecting to MySQL...');
    let connection;
    try {
        connection = await mysql.createConnection(DB_CONFIG);
        console.log('✅ Connected to MySQL\n');
    } catch (error) {
        console.error(`❌ Database connection failed: ${error.message}`);
        console.error('Make sure XAMPP MySQL is running!');
        process.exit(1);
    }

    try {
        // Get existing tags
        console.log('📑 Loading existing tags...');
        const [tagRows] = await connection.execute('SELECT id, slug FROM prompt_tags');
        const tagMap = {};
        tagRows.forEach(row => {
            tagMap[row.slug.toLowerCase()] = row.id;
        });
        console.log(`✅ Found ${Object.keys(tagMap).length} existing tags\n`);

        // Import prompts
        console.log('📥 Importing prompts...\n');
        let imported = 0;
        let skipped = 0;
        let errors = 0;

        for (const prompt of data.prompts) {
            try {
                // Skip if no prompt text
                if (!prompt.promptText || prompt.promptText.trim() === '') {
                    skipped++;
                    continue;
                }

                const caseNum = extractCaseNumber(prompt.title);

                // Check if already exists
                const [existing] = await connection.execute(
                    'SELECT id FROM prompts WHERE case_number = ?',
                    [caseNum]
                );

                if (existing.length > 0) {
                    skipped++;
                    continue;
                }

                // Get local thumbnail path
                const localThumbnail = prompt.thumbnail
                    ? getLocalImagePath(prompt.thumbnail, prompt.title, 0)
                    : '';

                // Insert prompt
                const [result] = await connection.execute(
                    `INSERT INTO prompts (case_number, title, prompt_text, source_url, thumbnail, prompt_count) 
                     VALUES (?, ?, ?, ?, ?, ?)`,
                    [
                        caseNum,
                        prompt.title,
                        prompt.promptText,
                        prompt.sourceUrl || '',
                        localThumbnail,
                        prompt.promptCount || 1
                    ]
                );

                const promptId = result.insertId;

                // Insert images with local paths
                if (prompt.images && Array.isArray(prompt.images)) {
                    const uniqueImages = [...new Set(prompt.images)];
                    for (let i = 0; i < uniqueImages.length; i++) {
                        const localImagePath = getLocalImagePath(uniqueImages[i], prompt.title, i);
                        await connection.execute(
                            'INSERT INTO prompt_images (prompt_id, image_url, display_order) VALUES (?, ?, ?)',
                            [promptId, localImagePath, i]
                        );
                    }
                }

                // Insert tag relations
                if (prompt.tags && Array.isArray(prompt.tags)) {
                    for (const tag of prompt.tags) {
                        // Skip Chinese tags and invalid ones
                        if (tag.includes('提示词') || tag.length > 30) continue;

                        const slug = tag.toLowerCase().trim();
                        let tagId = tagMap[slug];

                        // Create new tag if not exists
                        if (!tagId) {
                            try {
                                const [tagResult] = await connection.execute(
                                    'INSERT INTO prompt_tags (name, slug) VALUES (?, ?) ON DUPLICATE KEY UPDATE prompt_count = prompt_count',
                                    [tag, slug]
                                );
                                tagId = tagResult.insertId || tagResult.affectedRows;
                                if (tagResult.insertId) {
                                    tagMap[slug] = tagResult.insertId;
                                }
                            } catch (e) {
                                // Tag might already exist, try to get it
                                const [existingTag] = await connection.execute(
                                    'SELECT id FROM prompt_tags WHERE slug = ?',
                                    [slug]
                                );
                                if (existingTag.length > 0) {
                                    tagId = existingTag[0].id;
                                    tagMap[slug] = tagId;
                                }
                            }
                        }

                        if (tagId) {
                            try {
                                await connection.execute(
                                    'INSERT IGNORE INTO prompt_tag_relations (prompt_id, tag_id) VALUES (?, ?)',
                                    [promptId, tagId]
                                );
                            } catch (e) {
                                // Ignore duplicate key errors
                            }
                        }
                    }
                }

                imported++;

                // Progress update
                if (imported % 50 === 0) {
                    console.log(`  ✅ Imported ${imported} prompts...`);
                }

            } catch (error) {
                errors++;
                console.error(`  ❌ Error importing "${prompt.title}": ${error.message}`);
            }
        }

        // Update tag counts
        console.log('\n📊 Updating tag counts...');
        await connection.execute(`
            UPDATE prompt_tags pt SET prompt_count = (
                SELECT COUNT(*) FROM prompt_tag_relations ptr WHERE ptr.tag_id = pt.id
            )
        `);

        // Summary
        console.log('\n=================================');
        console.log('📊 IMPORT SUMMARY');
        console.log('=================================');
        console.log(`✅ Imported: ${imported}`);
        console.log(`⏭️  Skipped: ${skipped}`);
        console.log(`❌ Errors: ${errors}`);
        console.log(`📊 Total in file: ${data.prompts.length}`);

        // Show some stats
        const [countResult] = await connection.execute('SELECT COUNT(*) as count FROM prompts');
        const [imageCountResult] = await connection.execute('SELECT COUNT(*) as count FROM prompt_images');
        const [tagCountResult] = await connection.execute('SELECT COUNT(*) as count FROM prompt_tag_relations');

        console.log('\n📈 DATABASE STATS:');
        console.log(`   Prompts in DB: ${countResult[0].count}`);
        console.log(`   Images in DB: ${imageCountResult[0].count}`);
        console.log(`   Tag relations: ${tagCountResult[0].count}`);

    } catch (error) {
        console.error(`\n❌ Import error: ${error.message}`);
        throw error;
    } finally {
        await connection.end();
        console.log('\n🔌 Database connection closed');
    }

    console.log('\n🏁 Done!');
}

// Run
importData().catch(console.error);
