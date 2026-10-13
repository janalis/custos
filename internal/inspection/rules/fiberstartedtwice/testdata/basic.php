<?php
$f = new Fiber(fn() => Fiber::suspend()); $f->start(); <error descr="Start each fiber only once.">$f->start()</error>;
