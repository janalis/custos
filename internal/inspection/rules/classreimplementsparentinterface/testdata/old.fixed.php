<?php
namespace {
    // Polyfill for runtimes older than 8.0, where exceptions are not Stringable yet.
    if (\PHP_VERSION_ID < 80000) {
        interface Stringable { public function __toString(); }
    }
    class Failure extends \Exception implements \Stringable {}
    class Sized extends \ArrayObject {}
}
