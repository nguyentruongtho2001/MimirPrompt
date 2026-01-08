const PocketBase = require('pocketbase/cjs');

const PB_URL = 'http://127.0.0.1:8090';

async function main() {
    const pb = new PocketBase(PB_URL);
    await pb.admins.authWithPassword('admin@mimirprompt.com', 'mimirprompt123');
    console.log("🔍 Deep Debugging (Authenticated as Admin)...");

    // 1. Inspect Collection Schema
    try {
        const col = await pb.collections.getOne('prompts');
        console.log("\n📦 Collection Schema:");
        console.log(JSON.stringify(col.schema, null, 2));
    } catch (e) {
        console.error("❌ Failed to get collection:", e.message);
    }

    // 2. Inspect a record
    try {
        const list = await pb.collection('prompts').getList(1, 1);
        console.log(`\nPrompts Count: ${list.totalItems}`);
        if (list.items.length > 0) {
            const item = list.items[0];
            console.log("\n📄 Sample Record Full Dump:");
            console.log(JSON.stringify(item, null, 2));
        } else {
            console.log("⚠️ No prompts found.");
        }
    } catch (e) {
        console.error("❌ Failed to get record:", e.message);
    }

    // 2. Test Sort Params
    const sorts = ['id', '-created', 'created', 'updated', '-case_number', 'title', 'case_number'];
    console.log("\n🧪 Testing Sort Params:");

    for (const sortField of sorts) {
        process.stdout.write(`Testing sort='${sortField}'... `);
        try {
            await pb.collection('prompts').getList(1, 1, { sort: sortField });
            console.log("✅ OK");
        } catch (e) {
            console.log(`❌ FAILED: ${e.status} - ${e.message}`);
        }
    }
}

main();
