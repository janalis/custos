<?php
$f = new Fiber(fn() => 7); $f->start(); $f->getReturn();
$parameterized = new Fiber(fn($value) => Fiber::suspend());
$parameterized->start();
$parameterized->getReturn();
