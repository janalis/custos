<?php
$f=new Fiber(fn()=>1); $f->start(); $f->resume();
