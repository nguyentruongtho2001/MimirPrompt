/**
 * Reorder case numbers in descending order
 * Case 857 should be first (highest id), Case 1 should be last
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

    // Get all prompts ordered by ID
    const [prompts] = await connection.execute(`
        SELECT id, title, case_number FROM prompts ORDER BY id ASC
    `);

    console.log(`📊 Found ${prompts.length} prompts`);

    // Calculate new case numbers (857, 856, 855, ... 1)
    const totalPrompts = prompts.length;
    let updated = 0;

    for (let i = 0; i < prompts.length; i++) {
        const prompt = prompts[i];
        const newCaseNumber = totalPrompts - i; // 852, 851, 850, ...

        // Update case_number in database
        await connection.execute(
            'UPDATE prompts SET case_number = ? WHERE id = ?',
            [newCaseNumber, prompt.id]
        );

        // Also update the title if it contains "Case X:"
        let newTitle = prompt.title;
        const caseMatch = prompt.title.match(/^Case \d+:/);
        if (caseMatch) {
            newTitle = prompt.title.replace(/^Case \d+:/, `Case ${newCaseNumber}:`);
            await connection.execute(
                'UPDATE prompts SET title = ? WHERE id = ?',
                [newTitle, prompt.id]
            );
        }

        updated++;
        if (updated % 100 === 0) {
            console.log(`   Progress: ${updated}/${totalPrompts}`);
        }
    }

    console.log(`\n✅ Updated ${updated} prompts with new case numbers!`);

    // Show samples
    console.log('\n📝 Sample results:');
    const [samples] = await connection.execute(`
        SELECT id, case_number, title FROM prompts ORDER BY case_number DESC LIMIT 5
    `);
    samples.forEach(s => {
        console.log(`   ID ${s.id} → Case ${s.case_number}: ${s.title.substring(0, 40)}...`);
    });

    await connection.end();
}

main().catch(console.error);
