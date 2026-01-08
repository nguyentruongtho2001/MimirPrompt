/**
 * Translate Chinese prompts to English using Gemini API
 */

const mysql = require('mysql2/promise');
const { GoogleGenerativeAI } = require('@google/generative-ai');
const fs = require('fs');
const path = require('path');

// Configuration
const GEMINI_API_KEY = 'AIzaSyBi1wAGy7UMQB6nPORX4Ug28TaLPVMV5hQ';

const DB_CONFIG = {
    host: '127.0.0.1',
    port: 3306,
    user: 'root',
    password: '',
    database: 'mimirprompt_db',
    charset: 'utf8mb4'
};

// Delay between API calls (1 second to be safe)
const DELAY_MS = 1000;

// Progress file
const PROGRESS_FILE = path.join(__dirname, 'translate_gemini_progress.json');

// Initialize Gemini
const genAI = new GoogleGenerativeAI(GEMINI_API_KEY);
const model = genAI.getGenerativeModel({ model: 'gemini-2.0-flash-exp' });

/**
 * Check if text contains Chinese characters
 */
function containsChinese(text) {
    if (!text) return false;
    return /[\u4e00-\u9fff]/.test(text);
}

/**
 * Sleep function
 */
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Translate text using Gemini
 */
async function translateWithGemini(text, isTitle = false) {
    if (!text || !containsChinese(text)) {
        return text;
    }

    try {
        const prompt = isTitle
            ? `Translate this Chinese title to English. Keep it concise. Only return the translation, nothing else:\n\n${text}`
            : `Translate the following Chinese text to English. Keep the original formatting and structure. Only return the translation, nothing else:\n\n${text}`;

        const result = await model.generateContent(prompt);
        const response = await result.response;
        return response.text().trim();
    } catch (error) {
        console.error(`  ⚠️ Gemini error: ${error.message}`);

        // If rate limited, wait and retry
        if (error.message.includes('429') || error.message.includes('quota')) {
            console.log('  ⏳ Rate limited. Waiting 60s...');
            await sleep(60000);
            return translateWithGemini(text, isTitle);
        }

        return text; // Return original on failure
    }
}

/**
 * Translate title with case number handling
 */
async function translateTitle(title) {
    if (!title) return title;

    // Extract case number
    const caseMatch = title.match(/^案例\s*(\d+)[：:]\s*/);
    let casePrefix = '';
    let cleanTitle = title;

    if (caseMatch) {
        casePrefix = `Case ${caseMatch[1]}: `;
        cleanTitle = title.replace(/^案例\s*\d+[：:]\s*/, '');
    }

    if (!containsChinese(cleanTitle)) {
        return casePrefix + cleanTitle;
    }

    const translatedTitle = await translateWithGemini(cleanTitle, true);
    return casePrefix + translatedTitle;
}

/**
 * Load/Save progress
 */
function loadProgress() {
    try {
        if (fs.existsSync(PROGRESS_FILE)) {
            return JSON.parse(fs.readFileSync(PROGRESS_FILE, 'utf8')).lastId || 0;
        }
    } catch (e) { }
    return 0;
}

function saveProgress(lastId, count, total) {
    fs.writeFileSync(PROGRESS_FILE, JSON.stringify({
        lastId, count, total,
        timestamp: new Date().toISOString()
    }));
}

/**
 * Main function
 */
async function main() {
    console.log('🌐 MimirPrompt Translator (Gemini)');
    console.log('===================================\n');

    const lastId = loadProgress();
    if (lastId > 0) {
        console.log(`📂 Resuming from ID ${lastId}...\n`);
    }

    // Connect to database
    console.log('🔌 Connecting to MySQL...');
    let connection;
    try {
        connection = await mysql.createConnection(DB_CONFIG);
        console.log('✅ Connected to MySQL\n');
    } catch (error) {
        console.error(`❌ Database connection failed: ${error.message}`);
        process.exit(1);
    }

    try {
        // Get prompts with Chinese text
        console.log('📂 Finding prompts with Chinese text...');
        const [rows] = await connection.execute(
            'SELECT id, title, prompt_text FROM prompts WHERE id > ? ORDER BY id',
            [lastId]
        );

        const promptsToTranslate = rows.filter(row =>
            containsChinese(row.title) || containsChinese(row.prompt_text)
        );

        console.log(`✅ Found ${promptsToTranslate.length} prompts needing translation\n`);

        if (promptsToTranslate.length === 0) {
            console.log('🎉 All prompts are already translated!');
            if (fs.existsSync(PROGRESS_FILE)) fs.unlinkSync(PROGRESS_FILE);
            return;
        }

        console.log('🌐 Starting translation with Gemini...\n');

        // Handle Ctrl+C
        process.on('SIGINT', () => {
            console.log('\n\n⚠️ Interrupted! Progress saved.');
            process.exit(0);
        });

        const startTime = Date.now();
        let translated = 0;

        for (const row of promptsToTranslate) {
            try {
                let newTitle = row.title;
                let newPromptText = row.prompt_text;

                // Translate title
                if (containsChinese(row.title)) {
                    newTitle = await translateTitle(row.title);
                    await sleep(DELAY_MS);
                }

                // Translate prompt_text
                if (containsChinese(row.prompt_text)) {
                    // For very long prompts, translate in chunks
                    if (row.prompt_text.length > 10000) {
                        const chunks = row.prompt_text.split(/\n{2,}/);
                        const translatedChunks = [];

                        for (const chunk of chunks) {
                            if (containsChinese(chunk) && chunk.length > 10) {
                                translatedChunks.push(await translateWithGemini(chunk));
                                await sleep(DELAY_MS);
                            } else {
                                translatedChunks.push(chunk);
                            }
                        }
                        newPromptText = translatedChunks.join('\n\n');
                    } else {
                        newPromptText = await translateWithGemini(row.prompt_text);
                        await sleep(DELAY_MS);
                    }
                }

                // Update database
                await connection.execute(
                    'UPDATE prompts SET title = ?, prompt_text = ? WHERE id = ?',
                    [newTitle, newPromptText, row.id]
                );

                translated++;
                saveProgress(row.id, translated, promptsToTranslate.length);

                // Progress
                const elapsed = (Date.now() - startTime) / 1000;
                const rate = translated / elapsed * 60;
                const eta = (promptsToTranslate.length - translated) / rate;

                const titlePreview = newTitle.substring(0, 60);
                console.log(`✅ ${translated}/${promptsToTranslate.length}: ${titlePreview}...`);
                console.log(`   ⏱️  ${rate.toFixed(1)}/min | ETA: ${eta.toFixed(0)} min\n`);

            } catch (error) {
                console.error(`❌ Error ID ${row.id}: ${error.message}`);
                saveProgress(row.id - 1, translated, promptsToTranslate.length);
            }
        }

        // Cleanup
        if (fs.existsSync(PROGRESS_FILE)) fs.unlinkSync(PROGRESS_FILE);

        const totalTime = ((Date.now() - startTime) / 1000 / 60).toFixed(1);
        console.log('\n===================================');
        console.log('📊 TRANSLATION COMPLETE');
        console.log('===================================');
        console.log(`✅ Translated: ${translated}`);
        console.log(`⏱️  Total time: ${totalTime} minutes`);

    } finally {
        await connection.end();
        console.log('\n🔌 Database connection closed');
    }

    console.log('\n🏁 Done!');
}

main().catch(console.error);
