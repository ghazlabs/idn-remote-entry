# REST API

This system is expected to be called from Google App Script triggered upon form submission.

For authenticating the call, client is expected to submit predefined API key in `X-Api-Key` header. You can lookup the value of this key from `CLIENT_API_KEY` environment variable.

**Table of contents:**

- [REST API](#rest-api)
  - [Submit Manual Vacancy](#submit-manual-vacancy)
  - [Submit URL Vacancy](#submit-url-vacancy)
  - [Submit Bulk Vacancies](#submit-bulk-vacancies)
  - [Approve Vacancy as Admin](#approve-vacancy-as-admin)
  - [Reject Vacancy as Admin](#reject-vacancy-as-admin)
  - [System Errors](#system-errors)

## Submit Manual Vacancy

POST: `/vacancies`

This endpoint is used to submit a *manual* vacancy for processing. A manual vacancy refers to a submission where the client manually fills out each field in the vacancy data.

A successful call indicates that the vacancy has been submitted to the system but has not yet been processed. Processing will be handled asynchronously.

If the vacancy submitter is not whitelisted, the vacancy will remain pending until it receives approval. See [Approve Vacancy as Admin](#approve-vacancy-as-admin)

**Headers:**

| Field          | Type   | Required | Description                                    |
| -------------- | ------ | -------- | ---------------------------------------------- |
| `X-Api-Key`    | String | Yes      | The API Key for authenticating the call.       |
| `Content-Type` | String | Yes      | The only accepted value is `application/json`. |

**Body Payload:**

| Field               | Type           | Required | Description                                                                             |
| ------------------- | -------------- | -------- | --------------------------------------------------------------------------------------- |
| `submission_type`   | String         | Yes      | The value is `manual`.                                                                  |
| `submission_email`  | String         | No       | The email of submitter.                                                                 |
| `job_title`         | String         | Yes      | The title for the job (e.g Software Engineer, UI/UX Designer, Data Analyst, etc..).     |
| `company_name`      | String         | Yes      | The name of the company owned the vacancy (e.g Canonical, LeetCode, LaunchGood, etc..). |
| `company_location`  | String         | Yes      | The company HQ location where this job is offered (e.g London, UK).                     |
| `short_description` | String         | Yes      | The summary for the job description.                                                    |
| `relevant_tags`     | List of String | Yes      | The relevant tags for the job.                                                          |
| `apply_url`         | String         | Yes      | URL for applying the job, user can also put their email here.                           |

> Note:
>
> if `submission_email` is provided, system will check if the email is in approved list. If not provided or not in approved list, the request will be pending until it receives approval by the admin.

**Example Call:**

```json
POST /vacancies
X-Api-Key: 1ba9d286-20d3-43a0-b31b-013486d375a0
Content-Type: application/json

{
  "submission_type": "manual",
  "submission_email": "admin@ghazlabs.com",
  "job_title": "FullStack Software Engineer",
  "company_name": "Zero Gravity",
  "company_location": "London, UK",
  "short_description": "Zero Gravity (zerogravity.co.uk) is a UK-based startup with a mission to help low-income students get into top universities and careers. \r\n\r\nWe are looking to expand our engineering team in 2025. It will be a fully remote role from anywhere in Indonesia, starting with a 6-month contract that can be extended to a year or more.",
  "relevant_tags": ["ruby on rails", "full stack development", "system design", "api integration", "startup"],
  "apply_url": "debbie@zerogravity.co.uk"
}
```

**Success Response:**

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "ok": true,
  "ts": 1735432224
}
```

[Back to Top](#rest-api)

## Submit URL Vacancy

POST: `/vacancies`

This endpoint is used to submit a *url* vacancy for processing. A url vacancy refers to a submission where the client just submit the URL and let the system filled the vacancy data.

A successful call indicates that the vacancy has been submitted to the system but has not yet been processed. Processing will be handled asynchronously.

If the vacancy submitter is not whitelisted, the vacancy will remain pending until it receives approval. See [Approve Vacancy as Admin](#approve-vacancy-as-admin)

**Headers:**

| Field          | Type   | Required | Description                                    |
| -------------- | ------ | -------- | ---------------------------------------------- |
| `X-Api-Key`    | String | Yes      | The API Key for authenticating the call.       |
| `Content-Type` | String | Yes      | The only accepted value is `application/json`. |

**Body Payload:**

| Field              | Type   | Required | Description                                      |
| ------------------ | ------ | -------- | ------------------------------------------------ |
| `submission_type`  | String | Yes      | The value is `url`.                              |
| `submission_email` | String | No       | The email of submitter.                          |
| `apply_url`        | String | Yes      | URL for applying the job, the value must be URL. |

> Note:
>
> if `submission_email` is provided, system will check if the email is in approved list. If not provided or not in approved list, the request will be pending until it receives approval by the admin.

**Example Call:**

```json
POST /vacancies
X-Api-Key: 1ba9d286-20d3-43a0-b31b-013486d375a0
Content-Type: application/json

{
  "submission_type": "url",
  "submission_email": "admin@ghazlabs.com",
  "apply_url": "https://job-boards.eu.greenhouse.io/invertase/jobs/4492621101"
}
```

**Success Response:**

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "ok": true,
  "ts": 1735432224
}
```

[Back to Top](#rest-api)

## Submit Bulk Vacancies

POST: `/vacancies`

This endpoint is used to submit bulk vacancies for processing. This is a batch submission where the client submits multiple vacancies in a single request. This intended to be used for crawler. The vacancies are expected to have job title, company name and apply URL.

For bulk submission, the system will always send approval to the admin. The system will not check if the email is in approved list. For each action on vacancy, the system will treat the vacancy as URL Submission and will not reply to the email.

This e

**Headers:**

| Field          | Type   | Required | Description                                    |
| -------------- | ------ | -------- | ---------------------------------------------- |
| `X-Api-Key`    | String | Yes      | The API Key for authenticating the call.       |
| `Content-Type` | String | Yes      | The only accepted value is `application/json`. |

**Body Payload:**

| Field              | Type                | Required | Description                                                                           |
| ------------------ | ------------------- | -------- | ------------------------------------------------------------------------------------- |
| `submission_type`  | String              | Yes      | The value is `bulk`.                                                                 |
| `submission_email` | String              | No       | The email of submitter.                                                              |
| `bulk_vacancies`   | Array of Vacancy    | Yes      | Array of vacancy objects, each containing `job_title`, `company_name`, and `apply_url`.   |

> Note:
>
> if `submission_email` is provided, system will check if the email is in approved list. If not provided or not in approved list, the request will be pending until it receives approval by the admin.

**Example Call:**

```json
POST /vacancies
X-Api-Key: 1ba9d286-20d3-43a0-b31b-013486d375a0
Content-Type: application/json

{
  "submission_type": "bulk",
  "submission_email": "crawler",
  "bulk_vacancies": [
    {
      "job_title": "Software Engineer",
      "company_name": "Haraj",
      "apply_url": "https://www.haraj.com/jobs/haraj-software-engineer"
    },
    {
      "job_title": "Machine Learning Engineer",
      "company_name": "IDNRemote",
      "apply_url": "https://idnremote.com"
    }
  ]
}
```

**Success Response:**

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "ok": true,
  "ts": 1735432224
}
```

[Back to Top](#rest-api)

## Approve Vacancy as Admin

GET: `/vacancies/approve`

This endpoint is used to approve a vacancy that needs an approval. The vacancy needs approval if submitter's email is not whitelisted. This will be called directly from URL button in the email notification. Calling this endpoint with responded message id will return 400 Bad Request.

A successful call indicates that the vacancy has been approved and ready to be processed. Processing will be handled asynchronously.

**Query Params:**

| Field        | Type   | Required | Description                                                           |
| ------------ | ------ | -------- | --------------------------------------------------------------------- |
| `data`       | String | Yes      | The value is `base64` of `JSON Web Token` consisting of vacancy data. |
| `message_id` | String | No       | Message id of the email.                                              |

**Example Call:**

```bash
GET /vacancies/approve?data=?data=ZXl...&message_id=tcath05dmq@idnremote.com
```

**Success Response:**

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "ok": true,
  "ts": 1742223231
}
```

[Back to Top](#rest-api)

## Reject Vacancy as Admin

GET: `/vacancies/reject`

This endpoint is used to reject a vacancy that needs an approval. This will be called directly from URL in the email notification. Calling this endpoint with responded message id will return 400 Bad Request.

A successful call indicates that the vacancy has been rejected and will be ignored by the system.

**Query Params:**

| Field        | Type   | Required | Description                                                           |
| ------------ | ------ | -------- | --------------------------------------------------------------------- |
| `data`       | String | Yes      | The value is `base64` of `JSON Web Token` consisting of vacancy data. |
| `message_id` | String | No       | Message id of the email.                                              |

**Example Call:**

```bash
GET /vacancies/reject?data=?data=ZXl...&message_id=tcath05dmq@idnremote.com
```

**Success Response:**

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "ok": true,
  "ts": 1742223231
}
```

[Back to Top](#rest-api)

## System Errors

This section tells the error possible returned by the system.

- Invalid API Key

  ```json
  HTTP/1.1 401 Unauthorized
  Content-Type: application/json

  {
    "ok": false,
    "err": "ERR_INVALID_API_KEY",
    "msg": "invalid api key",
    "ts": 1735432224
  }
  ```

  This error indicates the submitted API Key in `X-Api-Key` header is invalid.

- Bad Request

  ```json
  HTTP/1.1 400 Bad Request
  Content-Type: application/json

  {
    "ok": false,
    "err": "ERR_BAD_REQUEST",
    "msg": "missing `apply_url`",
    "ts": 1735432224
  }
  ```

  This error indicates generic error on the request submitted by client. Please see the value of `msg` for details.

- Internal Server Error

  ```json
  HTTP/1.1 500 Internal Server Error
  Content-Type: application/json

  {
    "ok": false,
    "err": "ERR_INTERNAL_ERROR",
    "msg": "unable to connection to notion due: timeout",
    "ts": 1735432224
  }
  ```

  This error indicates generic error on server side. Please see the value of `msg` for details.

[Back to Top](#rest-api)
