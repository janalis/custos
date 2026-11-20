<?php // presets: per symfony; set: remove_unused_imports=true, single_import_per_statement=false, import_order=none
namespace N;

use A;
use X;

// kept comment
use Zed; /* note */ use function f;
foo(new A(), f());
use Y; /* trailing */ bar();
$x = $a ?:
    $b;
$y = $a ?
    : $b;
