# url-shortener

v0.0.3 : 24 June 2024 : deploy example to GCP

## Resolve short urls and redirect them

This web service uses a csv file to provide short urls which redirect to
remote urls. The server deals gracefully with calls to `/` and 404
errors with easily customizable web pages in `templates` and assets in
the `static` directory.

[Try it out on GCP](https://url-shortener-c35tmtbs2a-nw.a.run.app/)

The example csv file can be found [here](blob/main/data/short-urls.csv).

In development mode live reloading of the (minimal) web templates is
supported, and the remote urls are checked on startup.

In production mode the assets, including the csv file, are embedded into
the Go binary. A Dockerfile is included for easy deployment.

```
Usage:
  url-shortener 

A web server for redirecting short urls.

This uses a simple csv file of short,long urls as a database.

Run with the -d/-development flag to run in development mode, providing
live template reloads. In development mode, the urls are also checked at
startup.

Application Options:
  -i, --ipaddress=   ipaddress (default: 0.0.0.0)
  -p, --port=        port (default: 8000)
  -d, --development  run in development mode
  -t, --timeout=     development url checker timeout (default: 5s)
  -w, --workers=     development url checker workers (default: 8)

Help Options:
  -h, --help         Show this help message

```

Screenshot of the home page:

<img width="615" src="static/example.png" />

## Licence

Licensed under the [MIT Licence](LICENCE).
