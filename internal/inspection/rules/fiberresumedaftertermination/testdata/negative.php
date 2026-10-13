<?php
$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->resume();
$parameterized = new Fiber(fn($value) => 7);
$parameterized->start();
$parameterized->resume();
