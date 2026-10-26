<?php
namespace App\Tools {
    class Booter
    {
        public static function run(): void {}
    }
}

namespace {
    if (PHP_SAPI === 'cli') {
        \App\Tools\Booter::run();
    }
}
