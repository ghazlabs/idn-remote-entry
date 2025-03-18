# IDN Remote Entry

This system handles submissions from the [Submit Remote Job for Indonesian Talents](https://docs.google.com/forms/d/e/1FAIpQLSczxOnMSt-sK9X5e4tbccblbml0ik1r2fHKKCW-FST3hls5uQ/viewform?pli=1) form.

We opted to use Google Forms instead of creating a custom web app because it is much easier to set up and comes with built-in Google authentication. Since our primary goal is to simplify the process for the community to submit remote vacancies, this setup should be sufficient.

Below is the business logic for this system:

![Business Logic](./docs/business-logic.drawio.svg)

Below are the high-level architecture of this system:

![High Level Architecture](./docs/architecture.drawio.svg)

There are three main components in this system:

- *Server* => handle incoming request and delegate to vacancy queue
- *Vacancy Worker* => consume submitted vacancy queue, handle the extraction information from url vacancy and publish to notification queue. To extract the information from submitted url vacancy, will do the following:
  - Check if the url is has own parser registry
  - If the url is has own parser registry, use the parser registry to get html body and forward to LLM extract the information.
  - If the url is not has own parser registry, capture screenshot of the full page url web and forward to LLM to do the OCR.
- *Notification Worker* => consume notification queue and publish to channel

In production, we use Notion as storage and Whatsapp as channel. However in local development, we use local `.jsonl` as storage and email mailpit as channel.

## Getting Started

Before running this project, make sure you have Docker installed on your local machine and you need to set your own OpenAI API key as an environment variable. You can do this by running the following command:

```bash
export IDN_REMOTE_ENTRY_OPENAI_KEY=your_client_api_key
```

Then run the following command:

```bash
make run
```

The server will start running on `http://localhost:9864`.

Try to submit a url vacancy to the server by running the following command:

```bash
curl -X POST http://localhost:9864/vacancies \
  -H 'x-api-key: d2e97dca-1131-4344-af0a-b3406e7ecb06' \
  -d '{"submission_type": "url", "submission_email": "EMAIL_SUBMITTER", "apply_url": "URL_VACANCY"}'
```

> Note:
>
> Replace `EMAIL_SUBMITTER` with submitter's email to test the approval flow.
> Replace `URL_VACANCY` with the actual url vacancy.
> `x-api-key` is the api key to access the server you can find it in the [docker-compose](./deploy/local/run/docker-compose-local.yml) file.

If the `EMAIL_SUBMITTER` is in whitelisted email, the vacancy will be automatically processed without requiring approval. Otherwise, you will receive an approval notification via email on this url <http://localhost:8025>. If the vacancy requires approval, you can either approve it by clicking the link in the email or simply ignore it to reject. Once approved, the vacancy will be processing automatically.

> Note:
>
> Currently, the whitelisted emails are: `*@ghazlabs.com` and `*@idnremote.com`. You can edit this on `APPROVED_SUBMITTER_EMAILS` environment variable on `deploy/local/run/docker-compose-local.yml`.

When the vacancy is successfully processed, you will get the following events:

- The vacancy will be saved in local database which can be accessed by using this command in separate terminal: `make list-jobs`.
- You will get notification via email for the new vacancy on this url `http://localhost:8025` (in the production we used whatsapp not email).

For more detail checkout the REST API documentation for this system [here](./docs/rest_api.md).
