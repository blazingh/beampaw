
<p align="center">
  <a href="https://beampaw.xyz">
    <img src="https://raw.githubusercontent.com/blazingh/beampaw/main/public/beam_paw_icon.jpeg" alt="Logo" width=100 height=100>
  </a>

  <h3 align="center">Beam Paw</h3>

  <p align="center">
    Peer to Peer file transfer
    <br>
    <a href="https://github.com/blazingh/beampaw/issues/new?template=bug.md">Report bug</a>
    |
    <a href="https://github.com/blazingh/beampaw/issues/new?template=feature.md&labels=feature">Request feature</a>
  </p>
</p>


## Table of content

- [Quick start](#quick-start)
- [Run localy](#run-localy)
- [Self host](#self-host)


## Quick start


## Run Localy


1 - clone the repo and cd into the directory
```bash 
$ git clone https://github.com/blazingh/beampaw
$ cd beampaw
```
> you can run `$ make help` to see some quick helpful commands

2 - copy the example .env file and generate an ssh key file
```bash
cp example.env .env
$ ssh-keygen -t rsa -b 4096 -f id_rsa -q -N ''
```
3 - dowload npm dependicies and run tailwindcss build
```bash
$ npm install
$ npx tailwindcss -i ./styles.css -o ./public/index.css --minify
```
4 - start the project
```bash
$ go run main.go
```
> you can also use ` $ make run ` or `$ make run/watch` to run the project

**note :** if you want to develop the web front-end make sure to also run `$ npx tailwindcss -i ./styles.css -o ./public/index.css --minify` or `$ make tailwind-watch`

<br>


## Self Host
