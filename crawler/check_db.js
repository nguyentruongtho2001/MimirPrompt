const mysql = require('mysql2/promise');

async function check() {
    const connection = await mysql.createConnection({
        host: '127.0.0.1',
        user: 'root',
        password: '',
        database: 'mimirprompt_db'
    });

    console.log('\n📊 DATABASE VERIFICATION');
    console.log('========================\n');

    const [prompts] = await connection.execute('SELECT COUNT(*) as count FROM prompts');
    console.log(`✅ Prompts: ${prompts[0].count}`);

    const [images] = await connection.execute('SELECT COUNT(*) as count FROM prompt_images');
    console.log(`🖼️  Images: ${images[0].count}`);

    const [tags] = await connection.execute('SELECT COUNT(*) as count FROM prompt_tags');
    console.log(`🏷️  Tags: ${tags[0].count}`);

    const [relations] = await connection.execute('SELECT COUNT(*) as count FROM prompt_tag_relations');
    console.log(`🔗 Tag relations: ${relations[0].count}`);

    // Sample data
    console.log('\n📋 SAMPLE DATA (first 5 prompts):');
    console.log('----------------------------------');
    const [samples] = await connection.execute('SELECT id, case_number, title, thumbnail FROM prompts ORDER BY case_number DESC LIMIT 5');
    samples.forEach(p => {
        console.log(`  #${p.case_number}: ${p.title.substring(0, 40)}...`);
        console.log(`    Thumbnail: ${p.thumbnail}`);
    });

    await connection.end();
}

check().catch(console.error);
