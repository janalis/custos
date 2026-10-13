<?php
namespace App\Http {
    class Resp {}
}

namespace {
    // The alias name resolves to the original class in the index.
    class_alias(\App\Http\Resp::class, \App\Old\Resp::class);
    $r = \App\Old\Resp::class;
}
