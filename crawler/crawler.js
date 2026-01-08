/**
 * OpenNana Prompt Gallery Crawler - v3
 * Using exact DOM selectors from modal analysis
 */

const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

const BASE_URL = 'https://opennana.com/awesome-prompt-gallery/';
const OUTPUT_FILE = path.join(__dirname, '..', 'output', 'prompts.json');

async function crawlPrompts() {
    console.log('🚀 Starting OpenNana Prompt Crawler v3...');

    const browser = await chromium.launch({
        headless: true
    });

    const context = await browser.newContext({
        userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
    });

    const page = await context.newPage();

    try {
        console.log('📄 Navigating to gallery page...');
        await page.goto(BASE_URL, { waitUntil: 'domcontentloaded', timeout: 120000 });

        // Wait for prompt cards to load
        await page.waitForSelector('article.prompt-card', { timeout: 60000 });
        console.log('✅ Prompt cards detected!');

        // Wait for content to fully render
        await page.waitForTimeout(5000);

        // Scroll to load all items
        console.log('📜 Scrolling to load all items...');
        await autoScroll(page);
        await page.waitForTimeout(3000);

        // Get total count of cards
        const totalCards = await page.$$eval('article.prompt-card', cards => cards.length);
        console.log(`📊 Found ${totalCards} prompt cards on page`);

        // Extract all basic card data first
        console.log('🔍 Extracting all card data...');
        const allCardsData = await page.$$eval('article.prompt-card', cards => {
            return cards.map((card, index) => {
                const titleEl = card.querySelector('h3');
                const imgEl = card.querySelector('img');

                return {
                    index,
                    cardTitle: titleEl?.textContent?.trim() || `Prompt ${index + 1}`,
                    thumbnail: imgEl?.src || ''
                };
            });
        });

        console.log(`✅ Extracted ${allCardsData.length} cards`);

        // Now click on each card and get detailed data
        const fullPrompts = [];

        for (let i = 0; i < allCardsData.length; i++) {
            try {
                // Click on card
                await page.evaluate((idx) => {
                    const cards = document.querySelectorAll('article.prompt-card');
                    if (cards[idx]) {
                        cards[idx].click();
                    }
                }, i);

                // Wait for modal to open
                await page.waitForSelector('div.modal-content', { timeout: 5000 }).catch(() => null);
                await page.waitForTimeout(800);

                // Check if modal is open
                const hasModal = await page.$('div.modal-content');

                if (hasModal) {
                    // Scroll down in modal to see the full prompt
                    await page.evaluate(() => {
                        const modal = document.querySelector('div.modal-content');
                        if (modal) {
                            modal.scrollTop = modal.scrollHeight;
                        }
                    });

                    await page.waitForTimeout(500);

                    // Extract detailed data from modal-content
                    const detailData = await page.evaluate(() => {
                        const modal = document.querySelector('div.modal-content');
                        if (!modal) return null;

                        // Get title - from h1, h2, or h3 inside modal
                        let title = '';
                        const titleEl = modal.querySelector('h1, h2, h3');
                        if (titleEl) {
                            title = titleEl.textContent?.trim() || '';
                        }

                        // Get source URL
                        let sourceUrl = '';
                        const sourceLink = modal.querySelector('a[href*="http"]');
                        if (sourceLink) {
                            sourceUrl = sourceLink.href;
                        }

                        // Get images from modal
                        const images = Array.from(modal.querySelectorAll('img'))
                            .map(img => img.src)
                            .filter(src => src && !src.includes('data:'));

                        // Get prompt texts - look for content after "提示词" labels
                        const prompts = [];
                        const modalHTML = modal.innerHTML;
                        const modalText = modal.innerText || '';

                        // Find all prompt sections
                        const promptMatches = modalText.match(/提示词\s*\d*[^]*?(?=提示词\s*\d|$)/g);
                        if (promptMatches) {
                            for (const match of promptMatches) {
                                // Remove the "提示词 N" prefix and "复制" button text
                                let promptText = match
                                    .replace(/提示词\s*\d*/g, '')
                                    .replace(/复制/g, '')
                                    .trim();

                                if (promptText.length > 20) {
                                    prompts.push(promptText);
                                }
                            }
                        }

                        // Alternative: Get all text content and extract the long text portions
                        if (prompts.length === 0) {
                            const allDivs = modal.querySelectorAll('div');
                            for (const div of allDivs) {
                                const text = div.innerText?.trim() || '';
                                // Look for text that looks like a prompt (contains brackets, long enough)
                                if (text.length > 100 &&
                                    (text.includes('[') || text.includes('a ') || text.includes('an ')) &&
                                    !text.includes('提示词') &&
                                    !text.includes('复制')) {
                                    prompts.push(text);
                                    break;
                                }
                            }
                        }

                        // Get tags - look for span elements that look like tags
                        const tagElements = modal.querySelectorAll('span');
                        const tags = Array.from(tagElements)
                            .map(el => el.textContent?.trim())
                            .filter(t => t && t.length > 1 && t.length < 30 && !t.includes('复制'));

                        return {
                            title,
                            sourceUrl,
                            images,
                            tags: [...new Set(tags)],
                            promptTexts: prompts,
                            promptText: prompts.join('\n\n---\n\n')
                        };
                    });

                    if (detailData) {
                        fullPrompts.push({
                            index: allCardsData[i].index,
                            title: detailData.title || allCardsData[i].cardTitle,
                            thumbnail: allCardsData[i].thumbnail,
                            sourceUrl: detailData.sourceUrl,
                            images: detailData.images,
                            tags: detailData.tags,
                            promptText: detailData.promptText,
                            promptCount: detailData.promptTexts?.length || 0
                        });
                    }

                    // Close modal by clicking close button or pressing Escape
                    await page.click('button#modalClose').catch(() => { });
                    await page.keyboard.press('Escape').catch(() => { });
                    await page.waitForTimeout(300);

                } else {
                    // No modal found, just save basic data
                    fullPrompts.push({
                        index: allCardsData[i].index,
                        title: allCardsData[i].cardTitle,
                        thumbnail: allCardsData[i].thumbnail,
                        sourceUrl: '',
                        images: [allCardsData[i].thumbnail],
                        tags: [],
                        promptText: '',
                        promptCount: 0,
                        error: 'Modal not found'
                    });
                }

                // Log progress
                if ((i + 1) % 10 === 0 || i === 0) {
                    const lastItem = fullPrompts[fullPrompts.length - 1];
                    const titlePreview = (lastItem.title || 'N/A').substring(0, 35);
                    const hasPrompt = lastItem.promptText ? '✓' : '✗';
                    console.log(`  ✅ ${i + 1}/${totalCards}: ${titlePreview}... (prompt: ${hasPrompt})`);
                }

                // Save intermediate results every 50 items
                if ((i + 1) % 50 === 0) {
                    saveResults(fullPrompts, totalCards);
                }

            } catch (error) {
                console.log(`  ⚠️ Error on item ${i}: ${error.message.substring(0, 40)}`);
                fullPrompts.push({
                    index: allCardsData[i].index,
                    title: allCardsData[i].cardTitle,
                    thumbnail: allCardsData[i].thumbnail,
                    error: error.message
                });

                // Try to recover
                await page.click('button#modalClose').catch(() => { });
                await page.keyboard.press('Escape').catch(() => { });
                await page.waitForTimeout(300);
            }
        }

        // Final save
        saveResults(fullPrompts, totalCards);

        console.log('\n✅ Crawling completed!');
        console.log(`📊 Total crawled: ${fullPrompts.length}`);

        // Stats
        const withPrompts = fullPrompts.filter(p => p.promptText && p.promptText.length > 0).length;
        console.log(`📊 Items with prompts: ${withPrompts}/${fullPrompts.length}`);

        return { totalFound: totalCards, crawledCount: fullPrompts.length, withPrompts };

    } catch (error) {
        console.error('❌ Error:', error);
        throw error;
    } finally {
        await browser.close();
        console.log('🏁 Browser closed');
    }
}

function saveResults(prompts, totalFound) {
    const output = {
        source: BASE_URL,
        crawledAt: new Date().toISOString(),
        totalFound,
        crawledCount: prompts.length,
        prompts
    };

    const outputDir = path.dirname(OUTPUT_FILE);
    if (!fs.existsSync(outputDir)) {
        fs.mkdirSync(outputDir, { recursive: true });
    }

    fs.writeFileSync(OUTPUT_FILE, JSON.stringify(output, null, 2));
    console.log(`💾 Saved ${prompts.length} items to ${OUTPUT_FILE}`);
}

async function autoScroll(page) {
    await page.evaluate(async () => {
        await new Promise((resolve) => {
            let totalHeight = 0;
            const distance = 500;
            let scrollAttempts = 0;
            const maxAttempts = 100;

            const timer = setInterval(() => {
                const scrollHeight = document.body.scrollHeight;
                window.scrollBy(0, distance);
                totalHeight += distance;
                scrollAttempts++;

                if (totalHeight >= scrollHeight || scrollAttempts >= maxAttempts) {
                    clearInterval(timer);
                    window.scrollTo(0, 0); // Scroll back to top
                    resolve();
                }
            }, 150);
        });
    });
}

// Run crawler
crawlPrompts()
    .then((result) => {
        console.log(`\n🎉 Done! Total: ${result.crawledCount}/${result.totalFound} (with prompts: ${result.withPrompts})`);
    })
    .catch((error) => {
        console.error('❌ Crawling failed:', error);
        process.exit(1);
    });
