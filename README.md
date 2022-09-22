# Fish-Telegram-Captcha-bot

A telegram bot running on vercel to verify if the user is a human.
## Workflow

https://user-images.githubusercontent.com/19994286/189467303-d35ac426-fa06-4eef-a9e6-b3e5a87b7caa.mp4

## How to use?

### Verification-After-Request Mode

#### Public group

1. Enable "Approve new members" in group settings.
   ![image](https://user-images.githubusercontent.com/19994286/191825558-b2a65a97-b492-4379-b181-d7489f02fed6.png)
2. Add the bot to the group.
3. Promote the bot to admin.
4. Edit the admin rights, minimum rights are recommended.
   ![image](https://user-images.githubusercontent.com/19994286/191825676-d36cdf6c-4d69-49b1-9d24-42477f4ba3a4.png)

#### Private group

1. Go to "Group Settings" -> "Manage Invite Links".
2. Create a new invite link with "Request Admin Approval" enabled.
   ![image](https://user-images.githubusercontent.com/19994286/191826357-d6660d6b-19a4-487b-99ed-e6913870e790.png)
3. Add the bot to the group.
4. Promote the bot to admin.
5. Edit the admin rights, minimum rights are recommended.
   ![image](https://user-images.githubusercontent.com/19994286/191826424-ffa45df4-d2a0-4673-a61b-47249f029966.png)

**Notice: DO NOT Use the default invite link**

**There is a default invite link for a private groupchat in Telegram, which at the top of link list. The property of this invite link cannot be changes, which means that bot cannot verify the new member joined with this invite link.**

## Verification-After-Join Mode

In the groups linked with a channel, you can use this mode to verify new members.

1. Add the bot to the group.
2. Promote the bot to admin.
3. Edit the admin rights, minimum rights are recommended.
   ![image](https://user-images.githubusercontent.com/19994286/191827604-07372cbf-db1e-46f3-a7c2-50601630068a.png)

**Notice: The permission of "Invite Users via Link" CAN NOT be enabled on verification-after-join mode.**

**The bot will automatically handle the verification process. New members will be restricted after joining the group until they pass the CAPTCHA.**

## How to deploy?

### What you need:

- A vercel account
- The `token` of your telegram bot
- The `site-key` and `secret-key` of your hCaptcha

### Deploy to vercel

1. Create your telegram bot via [@BotFather](https://t.me/BotFather)
2. Go to the [Settings tab](https://dashboard.hcaptcha.com/settings) to get your secret key.
3. Go to the [Sites tab](https://dashboard.hcaptcha.com/sites) and create a new site key.
4. Fork this repo or click the button below to deploy it to vercel.
5. Go to the [Environment Variables](https://vercel.com/docs/environment-variables) tab and add the following variables:
    - `BOT_TOKEN`: The token of your telegram bot.
    - `SECRET_KEY`: The secret key of your hCaptcha account.
    - `SITE_KEY`: The site key of your active site.
6. Redeploy the project to make the environment variables take effect.
7. Set the webhook by requesting `https://api.telegram.org/bot[BOT_TOKEN]/setWebhook?url=https://[YOUR_DOMAIN]/webhook/[BOT_TOKEN]`

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/import/project?template=https://github.com/XiaoMengXinX/Fish-Telegram-Captcha-bot)

### Deploy to your own server

**Due to the limitations of vercel, verification-after-join mode can only work on the deployment on your own server.**

1. Deploy the project to vercel, but skip the step of setting the webhook.
2. Get the latest build on [GitHub Action](https://github.com/XiaoMengXinX/Fish-Telegram-Captcha-bot/actions).
3. Copy the url which assigned to your Production Deployment on vercel.
4. Run the bot on your server:
    ```bash
    FRONTEND_URL="YOUR_DEPLOYMENT_URL" \
    BOT_TOKEN="YOUR_BOT_TOKEN" \
    ./Fish-Telegram-Captcha-Bot
    ```
   