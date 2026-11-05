<?php
namespace First {
    \strlen('a');
}
namespace {
    strlen('a');
    array_map('strlen', ['a']);
    echo PHP_EOL;
}
namespace Last {
    \strlen('b');
}
