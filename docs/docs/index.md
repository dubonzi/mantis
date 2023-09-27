# About

<div align="center">
  <img src="img/logo.svg" width="190">
  <br/>
  <img src="img/name.svg" width="220">
</div>

<br/> 

<div align="center" class="badge">
  <a target="_blank" href="https://github.com/dubonzi/mantis/actions/workflows/go-test.yml">
    <img class="badge-item" src="https://github.com/dubonzi/mantis/actions/workflows/go-test.yml/badge.svg"/>
  </a>
  
  <a target="_blank" href="https://codecov.io/gh/dubonzi/mantis">
    <img class="badge-item" src="https://codecov.io/gh/dubonzi/mantis/graph/badge.svg?token=OJ97WK5VJJ"/>
  </a>

  <a target="_blank" href="https://dubonzi.github.io/mantis">
    <img class="badge-item" src="https://img.shields.io/badge/Docs-%F0%9F%93%9A-azure"/>
  </a>

</div>

</br>

Mantis is a REST API mocking tool, enabling you to mock any type of request, which makes development and running tests easier by not having to call a real service your app depends on. It is inspired by Wiremock and inherits some features to it.

I had the idea for Mantis when running a stress test at work using Wiremock to mock responses from various dependencies and realising it wasn't performing well and used too much resources to handle the throughput needed for the test.

The goal for Mantis is to be a fast and less resource-hungry alternative to other REST mocking tools available.

## Features

As of now, Mantis has all of the core features needed to be used in almost any scenario. But I intend to add more features over time, and of course, any pull request and suggestions are welcome!

- Easily mock any type of REST request
- Regex support
- JSON Path support
- Support for a Wiremock's `Scenario`-like feature
- Simulate production environments more accurately by defining delays on responses
- More to come