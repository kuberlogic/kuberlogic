# How to run locally

1. Setup new webhook on the chargebee side (https://kuberlogic.com/docs/configuring/billing). Currently supported only `subscription_created` event.
2. Configure environment `CHARGEBEE_SITE` and `CHARGEBEE_KEY` variables (see Makefile for the details).
3. For the ability to run the webhook locally, you need to set up [ngrok](https://ngrok.com/download) or [webhookrelay](https://webhookrelay.com/v1/examples/receiving-webhooks-on-localhost.html)
4. Run the webhook locally: 
    ```shell
    make run
    ```
5. Trigger the webhook on the chargebee side (click `Test Webhook`) on the webhooks pages.