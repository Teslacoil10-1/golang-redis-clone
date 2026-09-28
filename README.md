<a id="readme-top"></a>

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![APACHE-2.0][license-shield]][license-url]

<br />
<div align="center">
  <h3 align="center">Redis clone</h3>

  <p align="center">
    a redis clone wrote in rust and go
    <br />
    <a href="https://github.com/Teslacoil10-1/golang-redis-clone/docs"><strong>Explore the docs »</strong></a>
    <br />
    <a href="https://github.com/Teslacoil10-1/golang-redis-clone/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    &middot;
    <a href="https://github.com/Teslacoil10-1/golang-redis-clone/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div>

<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li><a href="#features">Features</a></li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#scripts--commands">Scripts / Commands</a></li>
    <li><a href="#license">License</a></li>
  </ol>
</details>

## About The Project

a redis clone wrote in rust and go

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

* [![Go][Go.dev]][Go-url]
* [![Rust][Rust.org]][Rust-url]
* [![gRPC][gRPC.io]][gRPC-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Features

- GRPC instead of RESP
- rust for the multi part AOF
- go for the rest of the codebase

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Getting Started

### Installation

```
git clone https://github.com/Teslacoil10-1/golang-redis-clone

cd golang-redis-clone
make up
make cli // to enter the cli (optional)
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Usage

there is two main entry points

for the cli it is cmd/cli/main.go
for the actual service its cmd/server/main.go

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Scripts / Commands

```
make up
make cli
make down
make test
make build
make proto
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## License

APACHE-2.0

<p align="right">(<a href="#readme-top">back to top</a>)</p>


[contributors-shield]: https://img.shields.io/github/contributors/Teslacoil10-1/golang-redis-clone.svg?style=for-the-badge
[contributors-url]: https://github.com/Teslacoil10-1/golang-redis-clone/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/Teslacoil10-1/golang-redis-clone.svg?style=for-the-badge
[forks-url]: https://github.com/Teslacoil10-1/golang-redis-clone/network/members
[stars-shield]: https://img.shields.io/github/stars/Teslacoil10-1/golang-redis-clone.svg?style=for-the-badge
[stars-url]: https://github.com/Teslacoil10-1/golang-redis-clone/stargazers
[issues-shield]: https://img.shields.io/github/issues/Teslacoil10-1/golang-redis-clone.svg?style=for-the-badge
[issues-url]: https://github.com/Teslacoil10-1/golang-redis-clone/issues
[license-shield]: https://img.shields.io/github/license/Teslacoil10-1/golang-redis-clone.svg?style=for-the-badge
[license-url]: https://github.com/Teslacoil10-1/golang-redis-clone/blob/master/LICENSE
[Go.dev]: https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white
[Go-url]: https://go.dev/
[Rust.org]: https://img.shields.io/badge/Rust-000000?style=for-the-badge&logo=rust&logoColor=white
[Rust-url]: https://www.rust-lang.org/
[gRPC.io]: https://img.shields.io/badge/gRPC-244C5A?style=for-the-badge&logo=grpc&logoColor=white
[gRPC-url]: https://grpc.io/
