# REST API

This system is expected to be called from Google App Script triggered upon form submission.

For authenticating the call, client is expected to submit predefined API key in `X-Api-Key` header. You can lookup the value of this key from `CLIENT_API_KEY` environment variable.

**Table of contents:**

- [Submit Manual Vacancy](#submit-manual-vacancy)
- [Submit URL Vacancy](#submit-url-vacancy)
- [System Errors](#system-errors)

## Submit Manual Vacancy

POST: `/vacancies`

This endpoint is used to submit a *manual* vacancy for processing. A manual vacancy refers to a submission where the client manually fills out each field in the vacancy data.

A successful call indicates that the vacancy has been submitted to the system but has not yet been processed. Processing will be handled asynchronously.

**Headers:**

- `X-Api-Key`, String => The API Key for authenticating the call.
- `Content-Type`, String => The only accepted value is `application/json`.

**Body Payload:**

- `submission_type`, String => The value is `manual`.
- `job_title`, String => The title for the job (e.g Software Engineer, UI/UX Designer, Data Analyst, etc..).
- `company_name`, String => The name of the company owned the vacancy (e.g Canonical, LeetCode, LaunchGood, etc..).
- `company_location`, String => The company HQ location where this job is offered (e.g London, UK).
- `short_description`, String => The summary for the job description.
- `relevant_tags`, String => The relevant tags for the job separated by comma.
- `apply_url`, String => URL for applying the job, user can also put their email here.

**Example Call:**

```
POST /vacancies
X-Api-Key: 1ba9d286-20d3-43a0-b31b-013486d375a0
Content-Type: application/json

{
	"submission_type": "manual",
	"job_title": "FullStack Software Engineer",
	"company_name": "Zero Gravity",
	"company_location": "London, UK",
	"short_description": "Zero Gravity (zerogravity.co.uk) is a UK-based startup with a mission to help low-income students get into top universities and careers. \r\n\r\nWe are looking to expand our engineering team in 2025. It will be a fully remote role from anywhere in Indonesia, starting with a 6-month contract that can be extended to a year or more.",
	"relevant_tags": "ruby on rails, full stack development, system design, api integration, startup",
	"apply_url": "debbie@zerogravity.co.uk"
}
```

**Success Response:**

```
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

**Headers:**

- `X-Api-Key`, String => The API Key for authenticating the call.
- `Content-Type`, String => The only accepted value is `application/json`.

**Body Payload:**

- `submission_type`, String => The value is `url`.
- `apply_url`, String => URL for applying the job, the value must be URL.

**Example Call:**

```
POST /vacancies
X-Api-Key: 1ba9d286-20d3-43a0-b31b-013486d375a0
Content-Type: application/json

{
	"submission_type": "url",
	"apply_url": "https://job-boards.eu.greenhouse.io/invertase/jobs/4492621101"
}
```

**Success Response:**

```
HTTP/1.1 200 OK
Content-Type: application/json

{
	"ok": true,
	"ts": 1735432224
}
```

[Back to Top](#rest-api)

## System Errors

This section tells the error possible returned by the system.

- Invalid API Key

	```
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

	```
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

	```
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