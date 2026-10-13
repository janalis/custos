<?php
$f = new Fiber(fn() => Fiber::suspend()); $f->start(); <error descr="Read the fiber return after termination.">$f->getReturn()</error>;
