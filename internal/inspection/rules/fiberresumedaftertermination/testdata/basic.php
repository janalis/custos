<?php
$f = new Fiber(fn() => 7); $f->start(); <error descr="Resume a fiber before it terminates.">$f->resume()</error>;
