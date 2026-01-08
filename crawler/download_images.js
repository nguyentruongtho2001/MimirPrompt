/**
 * Download All Images from prompts.json
 * Author: MimirPrompt Crawler
 * Description: Downloads all thumbnail and images from the crawled prompts data
 */

const fs = require('fs');
const path = require('path');
const https = require('https');
const http = require('http');

// Configuration
const DATA_FILE = path.join(__dirname, '..', 'data', 'prompts.json');
const OUTPUT_DIR = path.join(__dirname, '..', 'data', 'images');
const CONCURRENT_DOWNLOADS = 5; // Number of concurrent downloads
const RETRY_ATTEMPTS = 3;
const RETRY_DELAY = 2000; // ms

/**
 * Sanitize filename - remove invalid characters and limit length
 */
function sanitizeFilename(name, maxLength = 50) {
    if (!name) return 'untitled';
    
    // Extract case number from title (e.g., "案例 857：..." -> "857")
    const caseMatch = name.match(/案例\s*(\d+)/);
    const caseNum = caseMatch ? caseMatch[1] : '';
    
    // Remove case number prefix and special characters
    let cleaned = name
        .replace(/案例\s*\d+[：:]\s*/g, '') // Remove "案例 XXX："
        .replace(/[<>:"/\\|?*\x00-\x1F]/g, '') // Remove invalid Windows filename chars
        .replace(/\s+/g, '_') // Replace spaces with underscores
        .replace(/_{2,}/g, '_') // Replace multiple underscores
        .trim();
    
    // Limit length
    if (cleaned.length > maxLength) {
        cleaned = cleaned.substring(0, maxLength);
    }
    
    // Combine case number with cleaned title
    return caseNum ? `${caseNum}_${cleaned}` : cleaned;
}

/**
 * Get file extension from URL
 */
function getExtension(url) {
    const match = url.match(/\.(\w+)(?:\?|$)/);
    if (match) {
        return match[1].toLowerCase();
    }
    return 'jpg'; // Default to jpg
}

/**
 * Download a single file with retry logic
 */
function downloadFile(url, destPath, attempt = 1) {
    return new Promise((resolve, reject) => {
        // Skip if file already exists
        if (fs.existsSync(destPath)) {
            console.log(`  ⏭️  Already exists: ${path.basename(destPath)}`);
            resolve({ skipped: true });
            return;
        }

        const protocol = url.startsWith('https') ? https : http;
        
        const request = protocol.get(url, { timeout: 30000 }, (response) => {
            // Handle redirects
            if (response.statusCode === 301 || response.statusCode === 302) {
                const redirectUrl = response.headers.location;
                if (redirectUrl) {
                    downloadFile(redirectUrl, destPath, attempt)
                        .then(resolve)
                        .catch(reject);
                    return;
                }
            }

            if (response.statusCode !== 200) {
                if (attempt < RETRY_ATTEMPTS) {
                    setTimeout(() => {
                        downloadFile(url, destPath, attempt + 1)
                            .then(resolve)
                            .catch(reject);
                    }, RETRY_DELAY);
                    return;
                }
                reject(new Error(`HTTP ${response.statusCode} for ${url}`));
                return;
            }

            const fileStream = fs.createWriteStream(destPath);
            response.pipe(fileStream);

            fileStream.on('finish', () => {
                fileStream.close();
                resolve({ downloaded: true });
            });

            fileStream.on('error', (err) => {
                fs.unlink(destPath, () => {}); // Delete partial file
                reject(err);
            });
        });

        request.on('error', (err) => {
            if (attempt < RETRY_ATTEMPTS) {
                setTimeout(() => {
                    downloadFile(url, destPath, attempt + 1)
                        .then(resolve)
                        .catch(reject);
                }, RETRY_DELAY);
            } else {
                reject(err);
            }
        });

        request.on('timeout', () => {
            request.destroy();
            if (attempt < RETRY_ATTEMPTS) {
                downloadFile(url, destPath, attempt + 1)
                    .then(resolve)
                    .catch(reject);
            } else {
                reject(new Error('Timeout'));
            }
        });
    });
}

/**
 * Process downloads in batches
 */
async function processBatch(items, batchSize) {
    const results = { downloaded: 0, skipped: 0, failed: 0, errors: [] };
    
    for (let i = 0; i < items.length; i += batchSize) {
        const batch = items.slice(i, i + batchSize);
        const promises = batch.map(async (item) => {
            try {
                const result = await downloadFile(item.url, item.destPath);
                if (result.skipped) {
                    results.skipped++;
                } else {
                    results.downloaded++;
                    console.log(`  ✅ Downloaded: ${path.basename(item.destPath)}`);
                }
            } catch (error) {
                results.failed++;
                results.errors.push({ url: item.url, error: error.message });
                console.log(`  ❌ Failed: ${item.url} - ${error.message}`);
            }
        });
        
        await Promise.all(promises);
        
        // Progress update
        const progress = Math.min(i + batchSize, items.length);
        console.log(`\n📊 Progress: ${progress}/${items.length} images processed`);
    }
    
    return results;
}

/**
 * Main function
 */
async function main() {
    console.log('🚀 MimirPrompt Image Downloader');
    console.log('================================\n');

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

    // Create output directory
    if (!fs.existsSync(OUTPUT_DIR)) {
        fs.mkdirSync(OUTPUT_DIR, { recursive: true });
        console.log(`📁 Created output directory: ${OUTPUT_DIR}\n`);
    }

    // Collect all images to download
    const downloadItems = [];
    let totalImages = 0;

    for (const prompt of data.prompts) {
        const baseName = sanitizeFilename(prompt.title);
        
        // Collect all images (images array includes thumbnail usually)
        const allImages = new Set();
        
        if (prompt.thumbnail) {
            allImages.add(prompt.thumbnail);
        }
        
        if (prompt.images && Array.isArray(prompt.images)) {
            prompt.images.forEach(img => allImages.add(img));
        }

        // Create download items
        let imgIndex = 0;
        for (const imageUrl of allImages) {
            if (!imageUrl || !imageUrl.startsWith('http')) continue;
            
            const ext = getExtension(imageUrl);
            const filename = imgIndex === 0 
                ? `${baseName}.${ext}` 
                : `${baseName}_${imgIndex + 1}.${ext}`;
            
            downloadItems.push({
                url: imageUrl,
                destPath: path.join(OUTPUT_DIR, filename),
                promptTitle: prompt.title
            });
            
            imgIndex++;
            totalImages++;
        }
    }

    console.log(`📊 Total unique images to download: ${totalImages}`);
    console.log(`💾 Output directory: ${OUTPUT_DIR}\n`);
    console.log('🔄 Starting downloads...\n');

    // Download all images
    const startTime = Date.now();
    const results = await processBatch(downloadItems, CONCURRENT_DOWNLOADS);
    const duration = ((Date.now() - startTime) / 1000).toFixed(1);

    // Summary
    console.log('\n================================');
    console.log('📊 DOWNLOAD SUMMARY');
    console.log('================================');
    console.log(`✅ Downloaded: ${results.downloaded}`);
    console.log(`⏭️  Skipped (already exists): ${results.skipped}`);
    console.log(`❌ Failed: ${results.failed}`);
    console.log(`⏱️  Duration: ${duration}s`);
    console.log(`📁 Images saved to: ${OUTPUT_DIR}`);

    // Log errors if any
    if (results.errors.length > 0) {
        const errorLogPath = path.join(OUTPUT_DIR, 'download_errors.log');
        const errorLog = results.errors.map(e => `${e.url}\n  Error: ${e.error}`).join('\n\n');
        fs.writeFileSync(errorLogPath, errorLog);
        console.log(`\n⚠️  Error log saved to: ${errorLogPath}`);
    }

    console.log('\n🏁 Done!');
}

// Run
main().catch(console.error);
