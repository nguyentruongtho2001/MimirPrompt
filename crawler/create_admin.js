const mysql = require('mysql2/promise');
const bcrypt = require('bcrypt');

async function createAdmin() {
    console.log('🔑 Creating admin user...');

    const connection = await mysql.createConnection({
        host: '127.0.0.1',
        port: 3306,
        user: 'root',
        password: '',
        database: 'mimirprompt_db'
    });

    try {
        // Check if admin exists
        const [rows] = await connection.execute('SELECT * FROM admin_users WHERE username = ?', ['admin']);

        if (rows.length > 0) {
            console.log('⚠️ Admin user already exists. Updating password...');
            // Hash the password
            const hashedPassword = await bcrypt.hash('admin123', 10);
            await connection.execute(
                'UPDATE admin_users SET password_hash = ? WHERE username = ?',
                [hashedPassword, 'admin']
            );
            console.log('✅ Password updated!');
        } else {
            // Create new admin
            const hashedPassword = await bcrypt.hash('admin123', 10);
            await connection.execute(
                'INSERT INTO admin_users (username, password_hash) VALUES (?, ?)',
                ['admin', hashedPassword]
            );
            console.log('✅ Admin user created!');
        }

        console.log('\n📋 Login credentials:');
        console.log('   Username: admin');
        console.log('   Password: admin123');

    } catch (error) {
        console.error('❌ Error:', error.message);
    } finally {
        await connection.end();
    }
}

createAdmin();
