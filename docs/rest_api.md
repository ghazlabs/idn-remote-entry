# REST API

This system is expected to be called from Google App Script triggered upon form submission.

For authenticating the call, client is expected to submit predefined API key in `X-Api-Key` header. You can lookup the value of this key from `CLIENT_API_KEY` environment variable.

**Table of contents:**

- [Submit Manual Vacancy](#submit-manual-vacancy)
- [Submit URL Vacancy](#submit-url-vacancy)
- [Approve Vacancy as Admin](#approve-vacancy-as-admin)
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

| Field              | Type   | Required | Description                                        |
| ------------------ | ------ | -------- | -------------------------------------------------- |
| `submission_type`  | String | Yes      | The value is `url`.                                |
| `submission_email` | String | No       | The email of submitter.                            |
| `apply_url`        | String | Yes      | URL for applying the job, the value must be URL.   |

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

## Approve Vacancy as Admin

GET: `/vacancies/approve`

This endpoint is used to approve a vacancy that needs an approval. The vacancy needs approval if submitter's email is not whitelisted.

A successful call indicates that the vacancy has been approved and ready to be processed. Processing will be handled asynchronously.

**Query Params:**

- `data`, String => The value is `base64` of `JSON Web Token` consists of vacancy data. This query is mandatory.

**Example Call:**

```bash
GET /vacancies/approve?data=?data=ZXlKaGJHY2lPaUpJVXpJMU5pSXNJblI1Y0NJNklrcFhWQ0o5LmV5SnlaWFJ5YVdWeklqb3dMQ0p6ZFdKdGFYTnphVzl1WDJWdFlXbHNJam9pWVcxbGMyVnJZV2x6Wld0aGQyRnVRR2R0WVdsc0xtTnZiU0lzSW5OMVltMXBjM05wYjI1ZmRIbHdaU0k2SW0xaGJuVmhiQ0lzSW5aaFkyRnVZM2tpT25zaWFtOWlYM1JwZEd4bElqb2lSblZzYkZOMFlXTnJJRk52Wm5SM1lYSmxJRVZ1WjJsdVpXVnlJaXdpWTI5dGNHRnVlVjl1WVcxbElqb2lXbVZ5YnlCSGNtRjJhWFI1SWl3aVkyOXRjR0Z1ZVY5c2IyTmhkR2x2YmlJNklreHZibVJ2Yml3Z1ZVc2lMQ0p6YUc5eWRGOWtaWE5qY21sd2RHbHZiaUk2SWxwbGNtOGdSM0poZG1sMGVTQW9lbVZ5YjJkeVlYWnBkSGt1WTI4dWRXc3BJR2x6SUdFZ1ZVc3RZbUZ6WldRZ2MzUmhjblIxY0NCM2FYUm9JR0VnYldsemMybHZiaUIwYnlCb1pXeHdJR3h2ZHkxcGJtTnZiV1VnYzNSMVpHVnVkSE1nWjJWMElHbHVkRzhnZEc5d0lIVnVhWFpsY25OcGRHbGxjeUJoYm1RZ1kyRnlaV1Z5Y3k0Z1hISmNibHh5WEc1WFpTQmhjbVVnYkc5dmEybHVaeUIwYnlCbGVIQmhibVFnYjNWeUlHVnVaMmx1WldWeWFXNW5JSFJsWVcwZ2FXNGdNakF5TlM0Z1NYUWdkMmxzYkNCaVpTQmhJR1oxYkd4NUlISmxiVzkwWlNCeWIyeGxJR1p5YjIwZ1lXNTVkMmhsY21VZ2FXNGdTVzVrYjI1bGMybGhMQ0J6ZEdGeWRHbHVaeUIzYVhSb0lHRWdOaTF0YjI1MGFDQmpiMjUwY21GamRDQjBhR0YwSUdOaGJpQmlaU0JsZUhSbGJtUmxaQ0IwYnlCaElIbGxZWElnYjNJZ2JXOXlaUzRpTENKeVpXeGxkbUZ1ZEY5MFlXZHpJanBiSW5KMVlua2diMjRnY21GcGJITWlMQ0ptZFd4c0lITjBZV05ySUdSbGRtVnNiM0J0Wlc1MElpd2ljM2x6ZEdWdElHUmxjMmxuYmlJc0ltRndhU0JwYm5SbFozSmhkR2x2YmlJc0luTjBZWEowZFhBaVhTd2lZWEJ3YkhsZmRYSnNJam9pWkdWaVltbGxRSHBsY205bmNtRjJhWFI1TG1OdkxuVnJJbjE5LkRzVExLeWNhVEVZU3laaWp3U1BEMXVMWjd4ZFA1VHBFX3dveTVJZ1NlOVk=
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
