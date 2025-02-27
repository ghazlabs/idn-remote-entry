# IDN Remote Entry

This system handles submissions from the [Submit Remote Job for Indonesian Talents](https://docs.google.com/forms/d/e/1FAIpQLSczxOnMSt-sK9X5e4tbccblbml0ik1r2fHKKCW-FST3hls5uQ/viewform?pli=1) form.

We opted to use Google Forms instead of creating a custom web app because it is much easier to set up and comes with built-in Google authentication. Since our primary goal is to simplify the process for the community to submit remote vacancies, this setup should be sufficient.

Checkout the REST API documentation for this system [here](./docs/rest_api.md).

Below is the business logic for this system:

![Business Logic](./docs/business-logic.drawio.svg)

Below are the high-level architecture of this system:

![High Level Architecture](./docs/architecture.drawio.svg)

## Getting Started

To run this project locally, make sure you have Docker installed on your local machine. Then run the following command:

```bash
make run
```

> Caution: This command will start download `minicpm-v:8b` 5.5GB of Ollama model, make sure you have enough disk space and RAM to run.
