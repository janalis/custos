<?php
$size  = $big ? 'L' : (<warning descr="Avoid nesting ternary operators; use if/else or extract a variable.">$mid ? 'M' : 'S'</warning>);
$mode  = $ro ? (<warning descr="Avoid nesting ternary operators; use if/else or extract a variable.">$admin ? 'view-all' : 'view'</warning>) : 'edit';
$state = ((<warning descr="Avoid nesting ternary operators; use if/else or extract a variable.">$on ? 1 : 0</warning>)) ? 'up' : 'down';
$name  = (<warning descr="Avoid nesting ternary operators; use if/else or extract a variable.">$nick ?: $first</warning>) ?: 'anon';
$name  = $nick ?: (<warning descr="Avoid nesting ternary operators; use if/else or extract a variable.">$first ?: 'anon'</warning>);
$deep  = $a ? (<warning descr="Avoid nesting ternary operators; use if/else or extract a variable.">$b ? (<warning descr="Avoid nesting ternary operators; use if/else or extract a variable.">$c ? 1 : 2</warning>) : 3</warning>) : 4;
