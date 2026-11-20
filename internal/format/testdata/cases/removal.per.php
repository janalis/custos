<?php
// presets: per symfony; set: remove_unused_imports=true, single_import_per_statement=false, import_order=none
namespace N;

use A;

// kept comment
/* note */ use function f;

foo(new A(), f());
/* trailing */ bar();
$x = $a
    ?: $b;
$y = $a ?
    : $b;
