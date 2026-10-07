<?php
foreach ($queue as <error descr="Empty destructuring pattern: PHP 7+ rejects this with a fatal error.">list</error>(, , )) {
    tick();
}
foreach ($queue as <error descr="Empty destructuring pattern: PHP 7+ rejects this with a fatal error.">[</error>]) {
    tick();
}
foreach ($queue as <error descr="Empty destructuring pattern: PHP 7+ rejects this with a fatal error.">[</error>[], list()]) {
    tick();
}
