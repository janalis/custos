<?php
declare(strict_types=1);

$engines = [
    <weak_warning descr="Use \Random\Engine\Mt19937::class instead of the class name string.">'Random\Engine\Mt19937'</weak_warning>,
    // the first import went before the code: a second one is not chained
    <weak_warning descr="Use \Random\Engine\Secure::class instead of the class name string.">'Random\Engine\Secure'</weak_warning>,
];
