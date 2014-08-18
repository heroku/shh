System Heuristics Herald (aka Shh)
----

Gathers and relays system metrics

## Install

    go get github.com/heroku/shh

## Environment Variables

These are set in the config package. To view take a look at the docs for that package:

    go doc github.com/heroku/shh/config

## Building Debs on Heroku

```bash
heroku apps:create freeformz-build-shh --buildpack git://github.com/kr/heroku-buildpack-go.git
git push heroku
heroku open
```

Wait for the deb to be available, download and do what you want with it.

## 'Local' Development

1. Obtain a Linux system (only really tested on Ubuntu ATM)
2. Install Go (version 1.3 works fine)
3. Set GOPATH [appropriately](http://golang.org/doc/code.html)
3. `go get github.com/heroku/shh`
4. cd $GOPATH/src/github.com/heroku/shh
5. go build
6. ./shh

hack away

## Contributing

The goal for shh is to be a stable, low footprint system metrics
poller, and we welcome contributions, feedback and bug reports to make
that happen.

We're currently focused on supporting GNU/Linux systems, since that's
where we're using shh, but are open to supporting other platforms
provided the low footprint nature is preserved.

Please file bug reports through
[Github Issues](https://github.com/heroku/shh/issues). If you'd like
to contribute changes, please fork and submit a pull request.

## License

Copyright 2013 - 2014, Edward Muller, and contributors

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
