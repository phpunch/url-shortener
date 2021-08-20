# Basic URL shortener

Application Functionalities

- [x] API Client can send a url and be returned a shortened URL.
- [x] API Client can specify an expiration time for URLs, expired URLs must return HTTP 410
- [x] Input URL should be validated and respond with error if not a valid URL
- [x] Regex based blacklist for URLs, you can set blacklist in validate/validate.go
- [x] Visiting the Shortened URLs must redirect to the original URL with a HTTP 302 redirect,
404 if not found.
- [x] Hit counter for shortened URLs (increment with every hit)
- [x] Admin api (requiring token) to list
○ Short Code
○ Full Url
○ Expiry (if any)
○ Number of hits
- [x] Above list can filter by Short Code and keyword on origin url.
- [x] Admin api to delete a URL (after deletion shortened URLs must return HTTP 410)
- [ ] BONUS: Add a caching layer to avoid repeated database calls on popular URLs


How to run a service

- serve a service via Docker, port will be available at 8080
```sh
docker-compose up --build
```