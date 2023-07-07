# Compass

`compass` is a lightweight RPC/gRPC client wrapping the cosmos-sdk based off of the [lens client](https://github.com/strangelove-ventures/lens/tree/main/client). It is intended for use as is, or as a building block for larger applications and allows for standalone cosmos-sdk services.

# Testing

1) start a simd environment
2) run `make test`

## Starting Simd

* To start with a fresh simd environment run `make reset-simd`.
* To start the simd environment run `make start-simd`

## Running Tests

1) Start simd
2) Execute tests `make test`