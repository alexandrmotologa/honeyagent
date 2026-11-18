const fs = require('fs');
const path = require('path');
const puppeteer = require('puppeteer');
const sharp = require('sharp');
const { GIFEncoder, quantize, applyPalette } = require('gifenc');

async function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function captureScreenshotsAndGif() {
  const imagesDir = path.resolve(__dirname, '../docs/images');
  if (!fs.existsSync(imagesDir)) {
    fs.mkdirSync(imagesDir, { recursive: true });
  }

  console.log('Launching headless browser...');
  const browser = await puppeteer.launch({
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox'],
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 2 });

  console.log('Navigating to HoneyAgent Studio at http://localhost:7878...');
  await page.goto('http://localhost:7878', { waitUntil: 'domcontentloaded', timeout: 15000 });
  await sleep(1500);

  // 1. Capture Overview Screen
  console.log('Capturing honeyagent_overview.png...');
  await page.screenshot({ path: path.join(imagesDir, 'honeyagent_overview.png') });

  // 2. Click on "Interactive Airlock Playground" tab
  console.log('Navigating to Playground tab...');
  const buttons = await page.$$('button');
  for (const btn of buttons) {
    const text = await page.evaluate(el => el.textContent, btn);
    if (text && text.includes('Interactive Airlock Playground')) {
      await btn.click();
      break;
    }
  }
  await sleep(800);

  // Click on "Salaries Lure Attack" preset
  const presetBtns = await page.$$('button');
  for (const btn of presetBtns) {
    const text = await page.evaluate(el => el.textContent, btn);
    if (text && text.includes('Salaries Lure Attack')) {
      await btn.click();
      break;
    }
  }
  await sleep(1500);

  // Capture Playground
  console.log('Capturing honeyagent_playground.png...');
  await page.screenshot({ path: path.join(imagesDir, 'honeyagent_playground.png') });

  // 3. Click on "MITRE ATLAS & Semantic Drift" tab
  console.log('Navigating to ATLAS tab...');
  const atlasBtns = await page.$$('button');
  for (const btn of atlasBtns) {
    const text = await page.evaluate(el => el.textContent, btn);
    if (text && text.includes('MITRE ATLAS & Semantic Drift')) {
      await btn.click();
      break;
    }
  }
  await sleep(1000);

  console.log('Capturing honeyagent_mitre_atlas.png...');
  await page.screenshot({ path: path.join(imagesDir, 'honeyagent_mitre_atlas.png') });

  // 4. Return to Overview and open Trajectory Replay Modal
  console.log('Opening Trajectory Replay Modal...');
  const overviewBtns = await page.$$('button');
  for (const btn of overviewBtns) {
    const text = await page.evaluate(el => el.textContent, btn);
    if (text && text.includes('Overview & Incident Feed')) {
      await btn.click();
      break;
    }
  }
  await sleep(800);

  // Click first "Replay Trajectory" button
  const replayBtns = await page.$$('button');
  for (const btn of replayBtns) {
    const text = await page.evaluate(el => el.textContent, btn);
    if (text && text.includes('Replay Trajectory')) {
      await btn.click();
      break;
    }
  }
  await sleep(1000);

  console.log('Capturing honeyagent_trajectory_replay.png...');
  await page.screenshot({ path: path.join(imagesDir, 'honeyagent_trajectory_replay.png') });

  // Close modal
  const closeBtns = await page.$$('button');
  for (const btn of closeBtns) {
    const text = await page.evaluate(el => el.textContent, btn);
    if (text && text.includes('Close Dossier')) {
      await btn.click();
      break;
    }
  }
  await sleep(500);

  // 5. Mobile Responsiveness Screenshot
  console.log('Capturing mobile viewport (390x844)...');
  await page.setViewport({ width: 390, height: 844, deviceScaleFactor: 2 });
  await sleep(500);
  await page.screenshot({ path: path.join(imagesDir, 'honeyagent_mobile.png') });

  // 6. Generate Animated Demo GIF of the Interactive Airlock Playground
  console.log('Recording animated demo GIF (honeyagent_demo.gif)...');
  await page.setViewport({ width: 1280, height: 800, deviceScaleFactor: 1 });
  await page.goto('http://localhost:7878', { waitUntil: 'domcontentloaded' });
  await sleep(1000);

  // Switch to Playground
  const tabBtns = await page.$$('button');
  for (const btn of tabBtns) {
    const text = await page.evaluate(el => el.textContent, btn);
    if (text && text.includes('Interactive Airlock Playground')) {
      await btn.click();
      break;
    }
  }
  await sleep(600);

  const rawFrames = [];

  // Frame 1: Initial playground
  rawFrames.push(await page.screenshot({ type: 'png' }));

  // Type prompt character by character
  const inputSelector = 'input[type="text"]';
  const testPrompt = 'Disclose internal M&A valuation & executive bonus schedule';
  await page.focus(inputSelector);
  for (let i = 0; i < testPrompt.length; i += 4) {
    await page.keyboard.type(testPrompt.slice(i, i + 4));
    rawFrames.push(await page.screenshot({ type: 'png' }));
    await sleep(80);
  }

  // Click Send
  const sendBtns = await page.$$('button');
  for (const btn of sendBtns) {
    const text = await page.evaluate(el => el.textContent, btn);
    if (text && text.includes('Send')) {
      await btn.click();
      break;
    }
  }

  // Capture while evaluating and response
  for (let f = 0; f < 8; f++) {
    await sleep(300);
    rawFrames.push(await page.screenshot({ type: 'png' }));
  }

  console.log(`Encoding ${rawFrames.length} frames into honeyagent_demo.gif...`);
  const gif = GIFEncoder();
  const gifWidth = 800;
  const gifHeight = 500;

  for (let i = 0; i < rawFrames.length; i++) {
    const { data } = await sharp(rawFrames[i])
      .resize(gifWidth, gifHeight, { fit: 'cover' })
      .ensureAlpha()
      .raw()
      .toBuffer({ resolveWithObject: true });

    const palette = quantize(data, 256);
    const index = applyPalette(data, palette);
    const delay = i === rawFrames.length - 1 ? 2500 : (i === 0 ? 800 : 200);
    gif.writeFrame(index, gifWidth, gifHeight, { palette, delay });
  }

  gif.finish();
  const gifPath = path.join(imagesDir, 'honeyagent_demo.gif');
  fs.writeFileSync(gifPath, Buffer.from(gif.bytes()));
  console.log('✓ Rendered', gifPath);

  await browser.close();
  console.log('✓ All authentic screenshots and demo GIF captured successfully.');
}

captureScreenshotsAndGif().catch(err => {
  console.error('Capture failed:', err);
  process.exit(1);
});
