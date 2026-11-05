<?php
namespace First {
    <weak_warning descr="Write '\strlen(...)' to allow compile-time binding.">strlen('a')</weak_warning>;
}
namespace {
    strlen('a');
    array_map('strlen', ['a']);
    echo PHP_EOL;
}
namespace Last {
    <weak_warning descr="Write '\strlen(...)' to allow compile-time binding.">strlen('b')</weak_warning>;
}
