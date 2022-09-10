# Fish-Telegram-Captcha-bot

A telegram bot running on vercel to verify if the user is a human.
## Workflow

https://user-images.githubusercontent.com/19994286/189465818-2174bcc4-fe74-4749-89cd-315ec4b0638a.mp4

## Deploy

### What you need:

- A vercel account
- The `token` of your telegram bot
- The `site-key` and `secret-key` of your hCaptcha

### Deploy to vercel

1. Create your telegram bot via [@BotFather](https://t.me/BotFather)
2. Go to the [Settings tab](https://dashboard.hcaptcha.com/settings) to get your secret key.
3. Go to the [Sites tab](https://dashboard.hcaptcha.com/sites) and create a new site key.
3. Fork this repo or click the button below to deploy it to vercel.
4. Go to the [Environment Variables](https://vercel.com/docs/environment-variables) tab and add the following variables:
    - `BOT_TOKEN`: The token of your telegram bot.
    - `SECRET_KEY`: The secret key of your hCaptcha account.
    - `SITE_KEY`: The site key of your active site.
5. Redeploy the project to make the environment variables take effect.
6. Set the webhook by requesting `https://api.telegram.org/bot[BOT_TOKEN]/setWebhook?url=https://[YOUR_DOMAIN]/webhook/[BOT_TOKEN]`

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/import/project?template=https://github.com/XiaoMengXinX/Fish-Telegram-Captcha-bot)
