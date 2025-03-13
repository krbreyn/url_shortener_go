An example URL Shorterner webserver implemented in Go.

Features:

- Accepts URLs to shorten via netcat/nc over TCP, such as `echo "{url}" | netcat server port`

- Generates random 6-char URL strings

- Serves redirects via HTTP
