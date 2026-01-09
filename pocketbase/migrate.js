require('dotenv').config();
const PocketBase = require('pocketbase/cjs');
const mysql = require('mysql2/promise');

// Configuration from environment variables
const PB_URL = process.env.PB_URL || 'http://127.0.0.1:8090';
const ADMIN_EMAIL = process.env.ADMIN_EMAIL;
const ADMIN_PASS = process.env.ADMIN_PASS;

const MYSQL_CONFIG = {
    host: process.env.MYSQL_HOST || '127.0.0.1',
    port: parseInt(process.env.MYSQL_PORT) || 3306,
    user: process.env.MYSQL_USER || 'root',
    password: process.env.MYSQL_PASS || '',
    database: process.env.MYSQL_DB || 'mimirprompt_db'
};


async function main() {
    console.log('🚀 Starting re-migration from MySQL...');

    // 1. Setup PocketBase
    const pb = new PocketBase(PB_URL);
    await pb.admins.authWithPassword(ADMIN_EMAIL, ADMIN_PASS);

    // 2. Connect MySQL
    let db;
    try {
        db = await mysql.createConnection(MYSQL_CONFIG);
        console.log('✅ Connected to MySQL');
    } catch (e) {
        console.error('❌ MySQL Connection Failed:', e.message);
        process.exit(1);
    }

    // 3. Create Collections (this will handle clearing old collections)
    await createCollections(pb);

    // 4. Migrate Data
    const idMap = {
        authors: {},
        categories: {},
        tags: {}
    };

    await migrateAuthors(db, pb, idMap);
    await migrateCategories(db, pb, idMap);
    await migrateTags(db, pb, idMap);
    await migratePrompts(db, pb, idMap);
    console.log('🎉 Migration from MySQL completed successfully!');
    process.exit(0);
}

// ... migrateFromJson code ... (skipped in replacement if not targeted)

async function createCollections(pb) {
    console.log('📦 Creating collections...');

    // 0. Delete Prompts FIRST (Reverse dependency)
    try {
        const existing = await pb.collections.getOne('prompts');
        await pb.collections.delete(existing.id);
        console.log(`   - Deleted old prompts`);
    } catch (e) { }

    const simpleCollections = [
        {
            name: 'authors',
            type: 'base',
            fields: [
                { name: 'name', type: 'text', required: true },
                { name: 'username', type: 'text' },
                { name: 'platform', type: 'text' },
                { name: 'profile_url', type: 'text' },
                { name: 'avatar_url', type: 'text' },
                { name: 'bio', type: 'text' },
                { name: 'prompt_count', type: 'number' }
            ]
        },
        {
            name: 'tags',
            type: 'base',
            fields: [
                { name: 'name', type: 'text', required: true },
                { name: 'slug', type: 'text' },
                { name: 'description', type: 'text' },
                { name: 'prompt_count', type: 'number' }
            ]
        }
    ];

    // 1. Create simple collections
    const colIds = {};
    for (const col of simpleCollections) {
        let record;
        try {
            const existing = await pb.collections.getOne(col.name);
            await pb.collections.delete(existing.id);
            console.log(`   - Deleted old ${col.name}`);
        } catch (e) { }

        try {
            record = await pb.collections.create(col);
            console.log(`   + Created ${col.name} (${record.id})`);
            colIds[col.name] = record.id;
        } catch (e) {
            console.error(`   ! Failed to create ${col.name}: ${e.message}`);
            process.exit(1);
        }
    }

    // 2. Create Categories
    let categoriesId;
    try {
        const existing = await pb.collections.getOne('categories');
        await pb.collections.delete(existing.id);
    } catch (e) { }

    try {
        const catCol = await pb.collections.create({
            name: 'categories',
            type: 'base',
            fields: [
                { name: 'name', type: 'text', required: true },
                { name: 'slug', type: 'text' },
                { name: 'description', type: 'text' },
                { name: 'icon', type: 'text' },
                { name: 'color', type: 'text' },
                { name: 'display_order', type: 'number' },
                { name: 'is_active', type: 'bool' }
            ]
        });
        categoriesId = catCol.id;
        console.log(`   + Created categories (${categoriesId})`);

        // Update to add parent relation
        await pb.collections.update(categoriesId, {
            fields: [
                ...catCol.fields,
                { name: 'parent', type: 'relation', collectionId: categoriesId, maxSelect: 1 }
            ]
        });
        console.log(`   + Updated categories schema (self-relation)`);

        colIds['categories'] = categoriesId;
    } catch (e) {
        console.error(`   ! Failed to create categories: ${e.message}`);
        process.exit(1);
    }

    console.log("   -> colIds for Prompts:", JSON.stringify(colIds, null, 2));

    try {
        const promptsCol = await pb.collections.create({
            name: 'prompts',
            type: 'base',
            fields: [
                { name: 'case_number', type: 'number' },
                { name: 'title', type: 'text', required: true },
                { name: 'prompt_text', type: 'text', max: 0 },
                { name: 'source_url', type: 'text' },
                { name: 'thumbnail', type: 'text' },
                { name: 'is_hidden', type: 'bool' },
                { name: 'view_count', type: 'number' },
                { name: 'prompt_count', type: 'number' },
                { name: 'author', type: 'relation', collectionId: colIds['authors'], maxSelect: 1 },
                { name: 'category', type: 'relation', collectionId: colIds['categories'], maxSelect: 1 },
                { name: 'tags', type: 'relation', collectionId: colIds['tags'] },
                { name: 'images_list', type: 'json' }
            ]
        });
        console.log(`   + Created prompts (${promptsCol.id})`);
    } catch (e) {
        console.error(`   ! Failed to create prompts: ${e.message}`);
        process.exit(1);
    }
}

async function migrateAuthors(db, pb, idMap) {
    console.log('👤 Migrating authors...');
    const [rows] = await db.query('SELECT * FROM authors');
    for (const row of rows) {
        try {
            const data = {
                name: row.name,
                username: row.username,
                platform: row.platform,
                profile_url: row.profile_url,
                avatar_url: row.avatar_url,
                bio: row.bio,
                prompt_count: row.prompt_count,
                // Preserve created_at? PocketBase handles created/updated automatically.
                // We can set 'created' and 'updated' if we want, but SDK might override.
                // Let's stick to content.
            };

            // Check if exists to avoid dupes?
            // PocketBase create returns the record.
            const record = await pb.collection('authors').create(data);
            idMap.authors[row.id] = record.id;
        } catch (e) {
            console.error(`   ! Error migrating author ${row.name}: ${e.message}`);
        }
    }
    console.log(`   - Migrated ${rows.length} authors`);
}

async function migrateCategories(db, pb, idMap) {
    console.log('Vi Migrating categories...');
    const [rows] = await db.query('SELECT * FROM categories ORDER BY parent_id ASC');
    // Order by parent_id ASC to ensure parents exist first (if IDs are ordered)

    // We might need a second pass for parent relations if child comes before parent
    // But typically parents are created first.

    for (const row of rows) {
        try {
            const data = {
                name: row.name,
                slug: row.slug,
                description: row.description,
                icon: row.icon,
                color: row.color,
                display_order: row.display_order,
                is_active: !!row.is_active,
            };

            if (row.parent_id && idMap.categories[row.parent_id]) {
                data.parent = idMap.categories[row.parent_id];
            }

            const record = await pb.collection('categories').create(data);
            idMap.categories[row.id] = record.id;
        } catch (e) {
            console.error(`   ! Error migrating category ${row.name}: ${e.message}`);
        }
    }
}

async function migrateTags(db, pb, idMap) {
    console.log('🏷️ Migrating tags...');
    const [rows] = await db.query('SELECT * FROM prompt_tags'); // OR tags? Code said prompt_tags
    // prompt_repository.go uses 'prompt_tags' table

    for (const row of rows) {
        const data = {
            name: row.name,
            slug: row.slug,
            description: row.description,
            prompt_count: row.prompt_count
        };
        const record = await pb.collection('tags').create(data);
        idMap.tags[row.id] = record.id;
    }
}

async function migratePrompts(db, pb, idMap) {
    console.log('📝 Migrating prompts...');
    const [prompts] = await db.query('SELECT * FROM prompts');

    // Get images
    const [images] = await db.query('SELECT * FROM prompt_images ORDER BY display_order');
    const imagesByPrompt = {};
    images.forEach(img => {
        if (!imagesByPrompt[img.prompt_id]) imagesByPrompt[img.prompt_id] = [];
        imagesByPrompt[img.prompt_id].push(img.image_url);
    });

    // Get prompt tags relations
    const [relations] = await db.query('SELECT * FROM prompt_tag_relations');
    const tagsByPrompt = {};
    relations.forEach(rel => {
        if (!tagsByPrompt[rel.prompt_id]) tagsByPrompt[rel.prompt_id] = [];
        if (idMap.tags[rel.tag_id]) {
            tagsByPrompt[rel.prompt_id].push(idMap.tags[rel.tag_id]);
        }
    });

    for (const row of prompts) {
        try {
            const data = {
                case_number: row.case_number,
                title: row.title,
                prompt_text: row.prompt_text,
                source_url: row.source_url,
                thumbnail: row.thumbnail,
                is_hidden: !!row.is_hidden,
                view_count: row.view_count,
                images_list: imagesByPrompt[row.id] || [], // Storing as JSON array of URLs
                // Relations
                author: idMap.authors[row.author_id],
                category: idMap.categories[row.category_id],
                tags: tagsByPrompt[row.id] || []
            };

            await pb.collection('prompts').create(data);
        } catch (e) {
            console.error(`   ! Error prompt ${row.id}: ${e.message}`, JSON.stringify(e.data));
        }
    }
    console.log(`   - Migrated ${prompts.length} prompts`);
}

main().catch(console.error);
