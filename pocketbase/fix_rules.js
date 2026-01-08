const PocketBase = require('pocketbase/cjs');

const PB_URL = 'http://127.0.0.1:8090';
const ADMIN_EMAIL = 'admin@mimirprompt.com';
const ADMIN_PASS = 'mimirprompt123';

async function main() {
    console.log('🔧 Fixing Collection API Rules...');

    const pb = new PocketBase(PB_URL);

    try {
        await pb.admins.authWithPassword(ADMIN_EMAIL, ADMIN_PASS);
    } catch (e) {
        console.error("❌ Login failed. Make sure PocketBase is running!");
        process.exit(1);
    }

    const collections = ['prompts', 'tags', 'categories', 'authors'];

    for (const name of collections) {
        try {
            const col = await pb.collections.getOne(name);

            // Allow public list and view
            // In PocketBase, "" means public, null means admin only
            await pb.collections.update(col.id, {
                listRule: "",
                viewRule: "",
            });

            console.log(`✅ Updated rules for: ${name}`);
        } catch (e) {
            console.error(`❌ Failed to update ${name}:`, e.message);
        }
    }
}

main();
