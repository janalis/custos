<?php
namespace App {
    class Account
    {
        public function Account() {}
    }
}

namespace {
    class Legacy
    {
        public function <error descr="Class 'Legacy' uses an old-style constructor; rename it to __construct.">Legacy</error>() {}
    }
}
