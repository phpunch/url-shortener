# Basic URL shortener

Application Functionalities

- [x] User can send a url and specify an expiration time for URLs
- [x] Regex based blacklist for URLs, you can set blacklist in validate/validate.go
- [x] User can visit the shorten URLs and redirect to the original URL.
- [x] Service always counts every hit for shortened URLs
- [x] Admin can see a list of short code, full url, expiry (if any) and number of hits.
- [x] Admin can also filter above list by short code and keyword on origin url.
- [x] Admin can delete a URL by short code
- [ ] Add a caching layer to avoid repeated database calls on popular URLs


How to run a service

- serve a service via Docker, port will be available at 8080

```sh
docker-compose up --build
```

- see all APIs by visiting <http://localhost:8080/swagger/index.html>

- tests all functions in controller pkg by

```sh
go test ./controller/
```