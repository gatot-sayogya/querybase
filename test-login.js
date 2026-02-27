const puppeteer = require('puppeteer');

(async () => {
  const browser = await puppeteer.launch({ headless: true });
  const page = await browser.newPage();
  
  page.on('console', msg => console.log('PAGE LOG:', msg.text()));
  page.on('response', async response => {
    if (response.url().includes('/api/v1/auth/login')) {
      console.log('LOGIN RESPONSE STATUS:', response.status());
      try {
        console.log('LOGIN RESPONSE BODY:', await response.json());
      } catch (e) {}
    }
  });

  try {
    await page.goto('http://localhost:3000/login', { waitUntil: 'networkidle2' });
    await page.type('input[type="text"]', 'admin');
    await page.type('input[type="password"]', 'admin123');
    await page.click('button[type="submit"]');
    
    await page.waitForNavigation({ waitUntil: 'networkidle2', timeout: 5000 }).catch(() => console.log('No navigation occurred'));
    
    // Output any error element text if present
    const errorText = await page.evaluate(() => {
      const el = document.querySelector('.text-red-500, .bg-red-50');
      return el ? el.innerText : null;
    });
    if (errorText) console.log('UI ERROR DISPLAYED:', errorText);
    
    console.log('CURRENT URL:', page.url());
  } catch (err) {
    console.error('TEST ERROR:', err);
  } finally {
    await browser.close();
  }
})();
