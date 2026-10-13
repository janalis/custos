<?php
$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->resume();
