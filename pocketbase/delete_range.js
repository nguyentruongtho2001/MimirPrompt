const PocketBase = require('pocketbase/cjs');

const PB_URL = 'http://127.0.0.1:8090';
const ADMIN_EMAIL = 'admin@mimirprompt.com';
const ADMIN_PASS = 'mimirprompt123';

async function main() {
    const pb = new PocketBase(PB_URL);
    await pb.admins.authWithPassword(ADMIN_EMAIL, ADMIN_PASS);

    console.log("🔍 Fetching all prompts to calculate range...");

    // Fetch all items with the same sort as Frontend (-case_number)
    const allPrompts = await pb.collection('prompts').getFullList({
        sort: '-case_number',
    });

    console.log(`📊 Total prompts: ${allPrompts.length}`);

    // Frontend pagination logic:
    // Page 1: 0-19
    // Page 26 start index: (26 - 1) * 20 = 500
    // Page 42 end index: 42 * 20 = 840

    const startPage = 26;
    const endPage = 42;
    const perPage = 20;

    const startIndex = (startPage - 1) * perPage;
    const endIndex = endPage * perPage; // slice end is exclusive, so this covers up to 839 (total 840 items)

    console.log(`🎯 Target Range: Page ${startPage} to ${endPage}`);
    console.log(`   Indices: ${startIndex} to ${endIndex}`);

    const targets = allPrompts.slice(startIndex, endIndex);

    if (targets.length === 0) {
        console.log("⚠️ No prompts found in this range.");
        return;
    }

    console.log(`⚠️  Ready to DELETE ${targets.length} prompts.`);
    console.log(`   First item to delete: ${targets[0].title} (Case: ${targets[0].case_number})`);
    console.log(`   Last item to delete: ${targets[targets.length - 1].title} (Case: ${targets[targets.length - 1].case_number})`);

    // Batch delete
    console.log("\n🗑️  Deleting...");
    let count = 0;
    for (const p of targets) {
        try {
            await pb.collection('prompts').delete(p.id);
            process.stdout.write(".");
            count++;
        } catch (e) {
            console.error(`\n❌ Failed to delete ${p.id}: ${e.message}`);
        }
    }

    console.log(`\n\n✅ Successfully deleted ${count} prompts.`);
}

main().catch(console.error);
